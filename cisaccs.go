package cisaccs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

func (a *CisAccount) OneCisExecuteSsh(host string, port int, cmds []string) error {
	// Проверка на корректность
	if !a.initated {
		return errors.New("create this struct by New command")
	}

	var cnd namedevs.CiscoNameDevs
	hstData, err := cnd.GetByHostName(a.cisFileName, host)
	if err != nil {
		return err
	}
	hstAccount, found := hostdata.GetHostAccountByGroupName(a.pwdFileName, hstData.Group)
	if !found {
		return errors.New("not found account")
	}

	// Debug print account info
	fmt.Printf("hstAccount: %v\n", hstAccount)

	// Настройка и подключение.
	device, err := netrasp.New(host,
		netrasp.WithDriver("ios"),
		netrasp.WithSSHPort(port),
		netrasp.WithUsernamePasswordEnableSecret(hstAccount.Username, hstAccount.Password, hstAccount.Secret),
	)
	if err != nil {
		return fmt.Errorf("unable to init config: %v", err)
	}
	ctx, cancelOpen := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelOpen()

	err = device.Dial(ctx)
	if err != nil {
		//fmt.Printf("unable to connect: %v\n", err)
		return fmt.Errorf("unable to connect: %v", err)
	}
	defer device.Close(context.Background())

	ctx, cancelEnable := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelEnable()

	err = device.Enable(ctx)
	if err != nil {
		return fmt.Errorf("unable to Enable command: %v", err)
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
		fmt.Print(output)
	}
	device.Close(ctx)

	return nil
}

func (a *CisAccount) MultiCisExecuteSsh(hosts []string, port int, cmds []string) error {
	//
	if !a.initated {
		return errors.New("create this struct by New command")
	}
	if (port <= 0) || (port > 65534) {
		return errors.New("ssh number port need > 0 and < 65534")
	}

	// Перебираем указанные хосты.
	for _, host := range hosts {
		a.OneCisExecuteSsh(host, port, cmds)
	}

	return nil
}

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
