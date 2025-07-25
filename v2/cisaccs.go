package cisaccs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ales999/cisaccs/v2/internal/hostdata"
	"github.com/ales999/cisaccs/v2/internal/namedevs"
	"github.com/ales999/cisaccs/v2/internal/netrasp"
	"github.com/ales999/cisaccs/v2/internal/utils"
)

var cisDebug bool

// Выводить больше отладочной информации о ходе работы
func SetMoreOutputConnectInfo(moreDebug bool) {
	cisDebug = moreDebug
}

type CisAccount struct {
	initated    bool   // Инициализировано через New(...)
	cisFileName string // Файл с именами хостов, с указанием группы (hosts.yaml)
	pwdFileName string // Файл с акаунтами и паролями, по группам (groups.yaml)
}

func NewCisAccount(cisFileName string, pwdFileName string) *CisAccount {

	return &CisAccount{
		initated:    true,
		cisFileName: cisFileName,
		pwdFileName: pwdFileName,
	}
}

// GetIfaceByHost - получить имя интерфейса хоста, если указан
func (a *CisAccount) GetIfaceByHost(host string) (string, error) {

	var retstr string
	// Проверка на корректность
	if !a.initated {
		return retstr, errors.New("create this struct by New command")
	}

	var cnd namedevs.CiscoNameDevs

	hstData, err := cnd.GetHostDataByHostName(a.cisFileName, host) // get new CiscoNameDevs struct
	if err != nil {
		return retstr, err
	}
	return hstData.Iface, nil
}

// OneCisExecuteSsh - выполнить набор команд на одном хосте.
func (a *CisAccount) OneCisExecuteSsh(hostName string, port int, cmds []string, connectTimeOut ...int) ([]string, error) {

	// Приведем имя хоста к прописным буквам
	hostName = strings.ToLower(strings.TrimSpace(hostName))
	if len(hostName) == 0 {
		return nil, errors.New("host name is empty")
	}

	var outs []string // результат работы выполнения на cisco
	// Если необязательный параметр не указан то будем использовать его
	var dialTimeout = 30
	if len(connectTimeOut) > 0 {
		dialTimeout = connectTimeOut[0]
	}

	// Проверка на корректность
	if !a.initated {
		return outs, errors.New("not create this struct by NewCisAccount func")
	}

	var cnd namedevs.CiscoNameDevs
	// Запросим данные о хосте по  его имени
	hstData, err := cnd.GetHostDataByHostName(a.cisFileName, hostName)
	if err != nil {
		return outs, err
	}
	// Запрос данных для авторизации на хосте по имени группы
	hstAccount, found := hostdata.GetHostAccountByGroupName(a.pwdFileName, hstData.Group)
	if !found {
		return outs, fmt.Errorf("error: not found account %s", hostName)
	}

	// Debug print account info
	if cisDebug {
		fmt.Printf("!Connect to host: %s (%v)", hostName, hstData.HostIp)
	}

	// Настройка и подключение.
	device, err := netrasp.New(hstData.HostIp,
		netrasp.WithDriver("ios"),
		netrasp.WithSSHPort(port),
		netrasp.WithDialTimeout(time.Duration(dialTimeout)*time.Second),
		netrasp.WithUsernamePasswordEnableSecret(hstAccount.Username, hstAccount.Password, hstAccount.Secret),
		netrasp.WithInsecureIgnoreHostKey(),
	)
	if err != nil {
		return outs, fmt.Errorf("unable to init config: %v", err)
	}
	ctx, cancelOpen := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelOpen()

	err = device.Dial(ctx)
	if err != nil {
		return outs, fmt.Errorf("unable to connect: %v", err)
	}
	defer device.Close(context.Background())

	ctx, cancelEnable := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelEnable()

	// TODO: If user is privilege 15 - not need enable
	err = device.Enable(ctx)
	if err != nil {
		return outs, fmt.Errorf("unable to Enable command: %v", err)
	}

	ctx, cancelRun := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelRun()
	/*
		if CisDebug {
			fmt.Println(" - Done")
		}
	*/
	for _, sendCommand := range cmds {
		output, err := device.Run(ctx, sendCommand)
		if err != nil {
			fmt.Printf("unable to run command: %v\n", err)
			continue
		}
		multiouts := utils.ConvMultiStrToArrayStr(output)
		outs = append(outs, multiouts...)
	}
	device.Close(ctx)

	return outs, nil
}

// MultiCisWithByGroupNameExecuteSsh - выполнить команды на группе хостов, в указанной группе
func (a *CisAccount) MultiCisWithByGroupNameExecuteSsh(groupName string, port int, cmds []string, connectTimeOut ...int) ([][]string, error) {
	var arrouts [][]string // Возвращаемый массив

	// Если необязательный параметр не указан то будем использовать его
	var dialTimeout = 30
	if len(connectTimeOut) > 0 {
		dialTimeout = connectTimeOut[0]
	}

	var cnd namedevs.CiscoNameDevs
	// Получаем список хостов что входят в указанную группу
	hostgrps, err := cnd.GetHostsByGroupName(a.cisFileName, groupName)
	if err != nil {
		return nil, err
	}
	// Выполняем команды одну за другой
	for _, hsttorun := range hostgrps {
		// Вызываем функцию выше
		rethst, err := a.OneCisExecuteSsh(hsttorun, port, cmds, dialTimeout)
		if err != nil { // Если один их хостов например недоступен это не повод прерывать работу на остальных
			// Вернем ошибку с именем хоста и что случилось
			errstr := hsttorun + " : " + err.Error()
			arrouts = append(arrouts, []string{errstr})

		} else {
			// Ошибок нет, сохраняем вывод.
			arrouts = append(arrouts, rethst)
		}
	}

	return arrouts, nil

}

// MultiCisExecuteSsh - выполнить набор команд на множестве хостов
func (a *CisAccount) MultiCisExecuteSsh(hosts []string, port int, cmds []string) ([]string, error) {

	var arrouts []string // Возвращаемый массив
	//
	if !a.initated {
		return arrouts, errors.New("create this struct by New command")
	}
	if (port <= 0) || (port > 65534) {
		return arrouts, errors.New("ssh number port need > 0 and < 65534")
	}

	// Перебираем указанные хосты.
	for _, host := range hosts {
		// Для каждого хоста выполняем набор команд
		rets, err := a.OneCisExecuteSsh(host, port, cmds)
		if err != nil {
			fmt.Println(err)
		}
		// Добавим массив полученный от 'OneCisExecuteSsh' в возвращаемый массив
		arrouts = append(arrouts, rets...)
	}

	return arrouts, nil
}

/*
// Test get hosts by goup name
func (a *CisAccount) GetTestGoups(groupName string) {

	// a.cisFileName,
	var cnd namedevs.CiscoNameDevs
	hosts, err := cnd.GetHostsByGroupName(a.cisFileName, groupName)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(`----`)
	fmt.Println(hosts)

}
*/
