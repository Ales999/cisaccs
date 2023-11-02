package cisaccs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ales999/cisaccs/internal/hostdata"
	"github.com/ales999/cisaccs/internal/namedevs"
	"github.com/ales999/cisaccs/internal/netrasp"
	"github.com/ales999/cisaccs/internal/utils"
)

var CisDebug bool

type CisAccount struct {
	initated    bool
	cisFileName string
	pwdFileName string
}

func NewCisAccount(cisFileName string, pwdFileName string) *CisAccount {

	return &CisAccount{
		initated:    true,        // Инициализировано через New(...)
		cisFileName: cisFileName, // Файл с именами хостов, с указанием группы
		pwdFileName: pwdFileName, // Файл с акаунтами и паролями, по группам
	}
}

// GetIfaceByHost - получить имя интерфейса, если указан
func (a *CisAccount) GetIfaceByHost(host string) (string, error) {

	var retstr string
	// Проверка на корректность
	if !a.initated {
		return retstr, errors.New("create this struct by New command")
	}

	var cnd namedevs.CiscoNameDevs

	hstData, err := cnd.GetByHostName(a.cisFileName, host) // get new CiscoNameDevs struct
	if err != nil {
		return retstr, err
	}
	return hstData.Iface, nil
}

// OneCisExecuteSsh - выполнить набор команд на одном хосте.
func (a *CisAccount) OneCisExecuteSsh(host string, port int, cmds []string) ([]string, error) {

	var outs []string // результат работы выполнения на cisco

	// Проверка на корректность
	if !a.initated {
		return outs, errors.New("not create this struct by NewCisAccount func")
	}

	var cnd namedevs.CiscoNameDevs
	hstData, err := cnd.GetByHostName(a.cisFileName, host)
	if err != nil {
		return outs, err
	}
	hstAccount, found := hostdata.GetHostAccountByGroupName(a.pwdFileName, hstData.Group)
	if !found {
		return outs, fmt.Errorf("error: not found account %s", host)
	}

	// Debug print account info
	if CisDebug {
		fmt.Printf("!Connect to host: %s (%v)", host, hstData.HostIp)
	}

	// Настройка и подключение.
	device, err := netrasp.New(hstData.HostIp,
		netrasp.WithDriver("ios"),
		netrasp.WithSSHPort(port),
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

	err = device.Enable(ctx)
	if err != nil {
		return outs, fmt.Errorf("unable to Enable command: %v", err)
	}

	ctx, cancelRun := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelRun()

	if CisDebug {
		fmt.Println(" - Done")
	}

	for _, sendCommand := range cmds {
		output, err := device.Run(ctx, sendCommand)
		if err != nil {
			fmt.Printf("unable to run command: %v\n", err)
			continue
		}
		mulouts := utils.ConvMultiStrToArrayStr(output)
		outs = append(outs, mulouts...)
	}
	device.Close(ctx)

	return outs, nil
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
