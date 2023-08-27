package namedevs

// CiscoNameDev содержит в себе имя устройства (cisco) и его параметры для подключения
//
// PS: Кроме данных с именами и паролем
type CiscoNameDev struct {
	NameDev string // Имя узла (hostname)
	Group   string // Имя Группы.
	HostIp  string // Ip данного узла.
}

// newCiscoNameDev  - вернуть ссылку на новый экземпляр структуры
func NewCiscoNameDev(NameDev string, Group string, HostIp string) *CiscoNameDev {
	return &CiscoNameDev{
		NameDev: NameDev,
		Group:   Group,
		HostIp:  HostIp,
	}
}

type CiscoNameDevs []CiscoNameDev

func (c *CiscoNameDevs) GetByHostName(fileName string, hostName string) *CiscoNameDev {

	return &CiscoNameDev{}

}
