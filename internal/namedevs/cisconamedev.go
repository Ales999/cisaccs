package namedevs

import (
	"fmt"
	"strings"

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
}

// newCiscoNameDev  - вернуть ссылку на новый экземпляр структуры
func newCiscoNameDev(namedev string, group string, hostip string) *CiscoNameDev {
	return &CiscoNameDev{
		NameDev: namedev,
		Group:   group,
		HostIp:  hostip,
	}
}

type CiscoNameDevs []CiscoNameDev

func (c *CiscoNameDevs) GetByHostName(cisFileName string, hostName string) (*CiscoNameDev, error) {

	// var kcis *koanf.Koanf // Конфигурация для Cisco (cis.yaml)
	kcis := koanf.New(".")

	if err := kcis.Load(file.Provider(cisFileName), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	var ciscoNameDevs []CiscoNameDev

	// for _, dev := range confDevs {
	dev := hostName
	var cnd = newCiscoNameDev(dev, kcis.String(dev+".group"), kcis.String(dev+".host"))
	ciscoNameDevs = append(ciscoNameDevs, *cnd)
	//}

	return &ciscoNameDevs[0], nil

}

func (c *CiscoNameDevs) GetHostsByGroupName(grpName string) []string {

	var ret []string

	for _, dev := range *c {
		if strings.EqualFold(dev.Group, grpName) {
			ret = append(ret, dev.NameDev)
		}
	}
	return ret

}
