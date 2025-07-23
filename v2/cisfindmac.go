package cisaccs

import (
	"strings"

	"github.com/ales999/cisaccs/v2/internal/utils"
)

func CisFindMac(maclines []string, fndmac string) (bool, []utils.MacLineData) {

	var retmacs []utils.MacLineData
	var status bool
	// Перебираем строки
	for _, line := range maclines {
		// Парсим строку
		macl := utils.ParseMacLine(line)
		if macl.GetLen() > 0 {
			if strings.EqualFold(fndmac, macl.GetMac()) {
				retmacs = append(retmacs, macl)
				status = true
			}
		}
	}
	if status {
		return true, retmacs
	} else {
		return false, []utils.MacLineData{}
	}
}
