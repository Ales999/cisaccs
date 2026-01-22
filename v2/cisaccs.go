package cisaccs

import (
	"context"
	"errors"
	"fmt"
	"net"
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
	cnd         *namedevs.CiscoNameDevs
}

// NewCisAccount - Создать новый объект.
//
// @Parameters:
//
//	cisFileName - Имя файла с именами хостов, с указанием группы и IP (hosts.yaml).
//	pwdFileName - Имя файла с описанием учетных данных для подключения к устройствам, привязанных к группам (groups.yaml).
func NewCisAccount(cisFileName string, pwdFileName string) *CisAccount {

	return &CisAccount{
		initated:    true,
		cisFileName: cisFileName,
		pwdFileName: pwdFileName,
	}
}

// GetIfaceByHost - получить имя интерфейса хоста, если указан
func (ca *CisAccount) GetIfaceByHost(host string) (string, error) {

	var retstr string
	// Проверка на корректность
	if !ca.initated {
		return retstr, errors.New("create this struct by New command")
	}

	var cnd namedevs.CiscoNameDevs
	if ca.cnd == nil {
		hstData, err := cnd.GetHostDataByHostName(ca.cisFileName, host) // get new CiscoNameDevs struct
		if err != nil {
			return retstr, err
		}
		ca.cnd = (*namedevs.CiscoNameDevs)(hstData)
	}
	return ca.cnd.Iface, nil
}

// OneCisExecuteSsh - выполнить набор команд на одном хосте.
func (ca *CisAccount) OneCisExecuteSsh(hostName string, port int, cmds []string, connectTimeOut ...int) ([]string, error) {

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
	if !ca.initated {
		return outs, errors.New("not create this struct by NewCisAccount func")
	}

	var cnd namedevs.CiscoNameDevs
	// Если имя устройства не совпадает с именем в кеше то запросим ещё раз.
	if ca.cnd == nil || !strings.EqualFold(ca.cnd.NameDev, hostName) {
		// Запросим данные о хосте по  его имени
		hstData, err := cnd.GetHostDataByHostName(ca.cisFileName, hostName)
		if err != nil {
			return outs, err
		}
		ca.cnd = (*namedevs.CiscoNameDevs)(hstData)
	}
	// Запрос данных для авторизации на хосте по имени группы
	hstAccount, found := hostdata.GetHostAccountByGroupName(ca.pwdFileName, ca.cnd.Group)
	if !found {
		return outs, fmt.Errorf("error: not found account %s", hostName)
	}

	// Debug print account info
	if cisDebug {
		fmt.Printf("!Connect to host: %s (%v)", hostName, ca.cnd.HostIp)
	}

	// Настройка и подключение.
	device, err := netrasp.New(ca.cnd.HostIp,
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
func (ca *CisAccount) MultiCisWithByGroupNameExecuteSsh(groupName string, port int, cmds []string, connectTimeOut ...int) ([][]string, error) {
	var arrouts [][]string // Возвращаемый массив

	// Если необязательный параметр не указан то будем использовать его
	var dialTimeout = 30
	if len(connectTimeOut) > 0 {
		dialTimeout = connectTimeOut[0]
	}

	var cnd namedevs.CiscoNameDevs
	// Получаем список хостов что входят в указанную группу
	hostgrps, err := cnd.GetHostsByGroupName(ca.cisFileName, groupName)
	if err != nil {
		return nil, err
	}
	// Выполняем команды одну за другой
	for _, hsttorun := range hostgrps {
		// Вызываем функцию выше
		rethst, err := ca.OneCisExecuteSsh(hsttorun, port, cmds, dialTimeout)
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
func (ca *CisAccount) MultiCisExecuteSsh(hosts []string, port int, cmds []string) ([]string, error) {

	var arrouts []string // Возвращаемый массив
	//
	if !ca.initated {
		return arrouts, errors.New("create this struct by New command")
	}
	if (port <= 0) || (port > 65534) {
		return arrouts, errors.New("ssh number port need > 0 and < 65534")
	}

	// Перебираем указанные хосты.
	for _, host := range hosts {
		// Для каждого хоста выполняем набор команд
		rets, err := ca.OneCisExecuteSsh(host, port, cmds)
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

type CND namedevs.CiscoNameDev // Тип аналогичный внутреннему типу CiscoNameDev

// Вернуть массив всех хостов
func (ca *CisAccount) GetHostsDataByHostName() ([]*CND, error) {
	// Загрузить и вернуть все хосты и их IP.
	hostsData, err := namedevs.GetHostsDataByHostName(ca.cisFileName)
	if err != nil {
		return nil, err
	}
	// Создадим в памяти массив для возврата данных хостов.
	result := make([]*CND, len(hostsData))
	// Конвертируем в публичный тип CND
	for i, item := range hostsData {
		result[i] = (*CND)(item)
	}

	return result, nil
}

/*
type HDP hostdata.HostData

func newHostData(hd hostdata.HostData) HDP {
	return HDP{
		Username: hd.Username,
		Password: hd.Password,
		Secret:   hd.Secret,
	}
}
*/

// Проверить что такая группа сушествует для подключения хостов и есть данные для подключения.
func (ca *CisAccount) HostByGroupExists(groupName string) bool {

	_, ret := hostdata.GetHostAccountByGroupName(ca.pwdFileName, groupName)

	return ret
}

// OneCisExecuteSshStepByStep - тестовая версия --> выполнить набор команд на одном хосте,
// вывести вывод команд в консоль и вернуть все вместе.
func (ca *CisAccount) OneCisExecuteSshStepByStep(
	hostName string,
	port int,
	cmds []string,
	flagUseCiscoWrire bool,
	connectTimeOut ...int,
) error {
	// Приведение имени хоста к нижнему регистру и проверка на пустоту
	hostName = strings.ToLower(strings.TrimSpace(hostName))
	if len(hostName) == 0 {
		return errors.New("host name is empty")
	}

	// Проверка порта
	if port <= 0 || port > 65535 {
		return errors.New("invalid port")
	}

	// Установка таймаута подключения
	var dialTimeout = 30
	if len(connectTimeOut) > 0 {
		dialTimeout = connectTimeOut[0]
	}

	// Проверка инициализации структуры
	if !ca.initated {
		return errors.New("not created this struct by NewCisAccount func")
	}

	// Получение данных о хосте
	var cnd namedevs.CiscoNameDevs
	hstData, err := cnd.GetHostDataByHostName(ca.cisFileName, hostName)
	if err != nil {
		return err
	}

	// Получение учётных данных
	hstAccount, found := hostdata.GetHostAccountByGroupName(ca.pwdFileName, hstData.Group)
	if !found {
		return fmt.Errorf("error: not found account %s", hostName)
	}

	// Настройка и подключение
	device, err := netrasp.New(
		hstData.HostIp,
		netrasp.WithDriver("ios"),
		netrasp.WithSSHPort(port),
		netrasp.WithDialTimeout(time.Duration(dialTimeout)*time.Second),
		netrasp.WithUsernamePasswordEnableSecret(hstAccount.Username, hstAccount.Password, hstAccount.Secret),
		netrasp.WithInsecureIgnoreHostKey(),
	)
	if err != nil {
		return fmt.Errorf("unable to init config: %v", err)
	}
	defer device.Close(context.Background())

	// Диалог с устройством
	ctx, cancelOpen := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelOpen()
	if err := device.Dial(ctx); err != nil {
		return fmt.Errorf("unable to connect: %v", err)
	}

	// Включение режима enable
	ctxEnbl, cancelEnable := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelEnable()
	if err := device.Enable(ctxEnbl); err != nil {
		return fmt.Errorf("unable to Enable command: %v", err)
	}

	// Выполнение команд
	ctxRun, cancelRun := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelRun()

	if len(cmds) > 0 {
		fmt.Println("--------------------------------------------------")
		smNum := len(hostName) + 18
		sm := strings.Repeat(" ", smNum)

		for _, sendCommand := range cmds {
			sendCommand = strings.TrimLeft(sendCommand, " ")
			fmt.Println(sm + sendCommand)
			output, err := device.Run(ctxRun, sendCommand)
			if err != nil {
				fmt.Printf("unable to run command %s: %v\n", sendCommand, err)
				continue
			}
			fmt.Print(output)
		}
	}

	// Сохранение конфигурации (если нужно)
	if flagUseCiscoWrire {
		fmt.Print("Выполняю сохранение конфигурации на самой cisco...")
		_, err = device.Run(ctxRun, "wr mem")
		if err != nil {
			fmt.Printf("\nError: unable to run command wr mem: %v\n", err)
			return fmt.Errorf("unable to save config: %v", err)
		}
		fmt.Println(" - Выполнено.")
	}

	// Выход из сессии
	ctxExit, cancelExit := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelExit()
	if _, err := device.Run(ctxExit, "exit"); err != nil {
		if errors.Is(err, net.ErrClosed) {
			fmt.Printf("unable to closed session: %v\n", err)
		}
	}

	fmt.Println("Выход из сессии выполнен")

	return nil
}
