package cisaccs

import (
	"strings"

	"github.com/ales999/cisaccs/v2/internal/utils"
)

type MacLineData = utils.MacLineData
type IntDescSVI = utils.IntDescSVI

func ParseIntDescSVI(lines []string) []IntDescSVI {
	return utils.ParseIntDescSVI(lines)
}

func CisGetAllMacEntries(maclines []string) []MacLineData {
	var ret []MacLineData
	for _, line := range maclines {
		m := utils.ParseMacLine(line)
		if m.GetLen() > 0 {
			ret = append(ret, m)
		}
	}
	return ret
}

func CisFindMac(maclines []string, fndmac string) (bool, []MacLineData) {
	var retmacs []utils.MacLineData
	var status bool
	for _, line := range maclines {
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
		return false, []MacLineData{}
	}
}
