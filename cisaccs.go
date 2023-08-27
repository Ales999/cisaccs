package cisaccs

import (
	"errors"
	"fmt"
	"strconv"

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

	// Open cisFile
	//hostdatas := hostdata.New(cisFileName)

	return &CisAccount{
		initated:    true,        // Инициализировано через New(...)
		cisFileName: cisFileName, // Файл с именами хостов, с указанием группы
		pwdFileName: pwdFileName, // Файл с акаунтами и паролями, по группам
	}
}

func (a *CisAccount) OneCisExecuteSsh(host string, cmds []string) error {
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

	fmt.Printf("hstAccount: %v\n", hstAccount)

	return nil
}

func (a *CisAccount) MultiCisExecuteSsh(hosts []string, cmds []string) error {
	//
	if !a.initated {
		return errors.New("create this struct by New command")
	}

	// Перебираем указанные хосты.
	for _, host := range hosts {
		a.OneCisExecuteSsh(host, cmds)
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
