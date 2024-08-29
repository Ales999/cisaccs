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
	NameDev   string // Имя узла (hostname)
	Group     string // Имя Группы.
	HostIp    string // Ip данного узла (management).
	HostExtIp string // Внешний IP
	Iface     string // Имя интерфейса которым подключен у вышестояшего коммутатора, если есть.
}

// newCiscoNameDev  - вернуть ссылку на новый экземпляр структуры
func newCiscoNameDev(namedev string, group string, hostip string, hostextip string, iface string) *CiscoNameDev {
	return &CiscoNameDev{
		NameDev:   namedev,
		Group:     group,
		HostIp:    hostip,
		HostExtIp: hostextip,
		Iface:     iface,
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
		kcis.String(hostName+".hoost_e"),
		kcis.String(hostName+".spb4face"),
	)

	return cnd, nil

}

/*
// GetHostsByGroupName - получить массив данного типа хостов относящийся к заданной группе
// Пока вроде не требуется
func (c *CiscoNameDevs) GetsByGroupName(cisFileName string, groupName string) ([]*CiscoNameDev, error) {

	var ret []*CiscoNameDev

	kcis := koanf.New(".")

	if err := kcis.Load(file.Provider(cisFileName), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}
	// Вернуть все имена хостов
	var hostLists = kcis.MapKeys("")
	// Если список хостов не пуст
	if len(hostLists) > 0 {
		// Бежим по найденным спискам имен хостов
		for _, hst := range hostLists {
			// Если у данного хоста группа искомая, то хост добавляем в результат
			if strings.EqualFold(kcis.String(hst+".group"), groupName) {
				//Создадим  новую структуру
				var cnd = newCiscoNameDev(
					hst,
					kcis.String(hst+".group"),
					kcis.String(hst+".host"),
					kcis.String(hst+".hoost_e"),
					kcis.String(hst+".spb4face"),
				)
				// И добавим ее в массив структур
				ret = append(ret, cnd)

			}
		}
	}
	return ret, nil

}
*/

// GetHostsByGroupName - получить список хостов относящийся к заданной группе
func (c *CiscoNameDevs) GetHostsByGroupName(cisFileName string, grpName string) ([]string, error) {

	var ret []string

	kcis := koanf.New(".")

	if err := kcis.Load(file.Provider(cisFileName), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}
	// Вернуть все имена хостов
	var hostLists = kcis.MapKeys("")
	// Если список хостов не пуст
	if len(hostLists) > 0 {
		// Бежим по найденным спискам имен хостов
		for _, hst := range hostLists {
			// Если у данного хоста группа искомая, то хост добавляем в результат
			if strings.EqualFold(kcis.String(hst+".group"), grpName) {
				ret = append(ret, hst)
			}
		}
	}
	return ret, nil

}
