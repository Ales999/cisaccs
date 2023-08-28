package cisaccs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ales999/cisaccs/internal/hostdata"
	"github.com/ales999/cisaccs/internal/namedevs"
	"github.com/ales999/cisaccs/internal/netrasp"
)

// temp struct
type Cisdata struct {
	Ndevs   []namedevs.CiscoNameDev
	Nstdats map[string]hostdata.HostData
}

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

func (a *CisAccount) OneCisExecuteSsh(host string, port int, cmds []string) ([]string, error) {

	var outs []string // результат работы выполнения на cisco

	// Проверка на корректность
	if !a.initated {
		return outs, errors.New("create this struct by New command")
	}

	var cnd namedevs.CiscoNameDevs
	hstData, err := cnd.GetByHostName(a.cisFileName, host)
	if err != nil {
		return outs, err
	}
	hstAccount, found := hostdata.GetHostAccountByGroupName(a.pwdFileName, hstData.Group)
	if !found {
		return outs, errors.New("not found account")
	}

	// Debug print account info
	//fmt.Printf("hstAccount: %v\n", hstAccount)
	fmt.Printf("Connect to %s (%v)\n", host, hstData.HostIp)

	// Настройка и подключение.
	device, err := netrasp.New(hstData.HostIp,
		netrasp.WithDriver("ios"),
		netrasp.WithSSHPort(port),
		netrasp.WithUsernamePasswordEnableSecret(hstAccount.Username, hstAccount.Password, hstAccount.Secret),
	)
	if err != nil {
		return outs, fmt.Errorf("unable to init config: %v", err)
	}
	ctx, cancelOpen := context.WithTimeoutCause(context.Background(), 30*time.Second, fmt.Errorf("open session timeout")) //.WithTimeout(context.Background(), 10*time.Second)
	defer cancelOpen()

	err = device.Dial(ctx)
	if err != nil {
		//fmt.Printf("unable to connect: %v\n", err)
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

	fmt.Println("--------------------------------------------------")
	for _, sendCommand := range cmds {
		output, err := device.Run(ctx, sendCommand)
		if err != nil {
			fmt.Printf("unable to run command: %v\n", err)
			continue
		}
		outs = append(outs, output)
	}
	device.Close(ctx)

	return outs, nil
}

func (a *CisAccount) MultiCisExecuteSsh(hosts []string, port int, cmds []string) ([]string, error) {
	var outs []string
	//
	if !a.initated {
		return outs, errors.New("create this struct by New command")
	}
	if (port <= 0) || (port > 65534) {
		return outs, errors.New("ssh number port need > 0 and < 65534")
	}

	// Перебираем указанные хосты.
	for _, host := range hosts {
		rets, err := a.OneCisExecuteSsh(host, port, cmds)
		if err != nil {
			fmt.Println(err)
		}
		for _, ret := range rets {
			outs = append(outs, ret)
		}

	}

	return outs, nil
}

/*
// test using internal using
func (a *CisAccount) Testme(hostIp string, port string, userName string, userPassword string) {

	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	// Настройка и подключение.
	device, err := netrasp.New(hostIp,
		netrasp.WithDriver("ios"),
		netrasp.WithSSHPort(portInt),
		netrasp.WithUsernamePassword(userName, userPassword),
	)
	if err != nil {
		panic(err)
	}

	// test print
	fmt.Println(device)

}
*/
