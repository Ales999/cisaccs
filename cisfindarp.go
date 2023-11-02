package cisaccs

import (
	"strings"

	"github.com/ales999/cisaccs/internal/utils"
)

// Парсинг ARP строк и поиск соответствия IP или MAC
func CisFindArp(arplines []string, ipormac string) (bool, utils.ArpLineData) {
	// Перебираем строки до первого совпадения.
	for _, line := range arplines {
		// Парсим строку
		arpl := utils.ParseArpLine(line)
		if arpl.GetLen() > 0 {
			if strings.EqualFold(arpl.GetIp(), ipormac) || strings.EqualFold(arpl.GetMac(), ipormac) {
				return true, arpl
			}
		}
	}
	return false, utils.ArpLineData{}
}
