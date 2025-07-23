package cisaccs

import (
	"github.com/ales999/cisaccs/v2/internal/utils"
)

// Парсинг строк вывода 'sh vlan br' и поиск нужного по ID
func CisFindVlan(vlanlines []string, fndvlandid int) (bool, utils.VlanLineData) {
	// Проверка параметров
	if len(vlanlines) > 0 && fndvlandid > 0 && fndvlandid < 4096 {
		// Бежим по массиву и как только найдем нужный ID - возврат с результатом
		for _, line := range vlanlines {
			vlan := utils.ParseVlan(line)
			if vlan.GetId() == fndvlandid {
				return true, vlan
			}
		}

	}
	return false, utils.VlanLineData{}
}

// Парсинг всех строк и возврат масива отпарсенных VLAN-ов
func CisGetVlans(vlanlines []string) []utils.VlanLineData {

	var retvlans []utils.VlanLineData
	// Проверка параметров
	if len(vlanlines) > 0 {
		// Бежим по всем строкам что cisco вернула.
		for _, line := range vlanlines {
			_nvl := utils.ParseVlan(line)
			// Проверка что запись не пустая
			if _nvl.GetId() == 0 {
				continue
			}
			// Если строка не пуста - добавляем
			retvlans = append(retvlans, _nvl)
		}
	}
	return retvlans
}
