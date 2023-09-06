package namedevs

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// CiscoNameDev содержит в себе имя устройства (cisco) и его параметры для подключения
//
// PS: Кроме данных с именами и паролем
type CiscoNameDev struct {
	NameDev string // Имя узла (hostname)
	Group   string // Имя Группы.
	HostIp  string // Ip данного узла.
	Iface   string // Имя интерфейса которым подключен у вышестояшего коммутатора, если есть.
}

// newCiscoNameDev  - вернуть ссылку на новый экземпляр структуры
func newCiscoNameDev(namedev string, group string, hostip string, iface string) *CiscoNameDev {
	return &CiscoNameDev{
		NameDev: namedev,
		Group:   group,
		HostIp:  hostip,
		Iface:   iface,
	}
}

type CiscoNameDevs CiscoNameDev

func (c *CiscoNameDevs) GetByHostName(cisFileName string, hostName string) (*CiscoNameDev, error) {

	// var kcis *koanf.Koanf // Конфигурация для Cisco (cis.yaml)
	kcis := koanf.New(".")

	if err := kcis.Load(file.Provider(cisFileName), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	//dev := hostName
	var cnd = newCiscoNameDev(
		hostName,
		kcis.String(hostName+".group"),
		kcis.String(hostName+".host"),
		kcis.String(hostName+".spb4face"),
	)

	//var ciscoNameDevs CiscoNameDev
	//ciscoNameDevs = append(ciscoNameDevs, *cnd)

	//return &ciscoNameDevs[0], nil
	return cnd, nil

}

/*
// TODO: add find by group
func (c *CiscoNameDevs) GetHostsByGroupName(grpName string) []string {

	var ret []string

	for _, dev := range *c {
		if strings.EqualFold(dev.Group, grpName) {
			ret = append(ret, dev.NameDev)
		}
	}
	return ret

}
*/
