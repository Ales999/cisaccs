package utils

import (
	"regexp"
	"strconv"
	"strings"
)

type IntDescSVI struct {
	Interface   string
	VlanID      int
	Status      string
	Protocol    string
	Description string
}

var svilineRe = regexp.MustCompile(`^\s*(Vl\d+)\s{2,}(\S+)\s{2,}(\S+)(?:\s+(.*))?`)

func ParseIntDescSVI(lines []string) []IntDescSVI {
	var ret []IntDescSVI
	for _, line := range lines {
		tr := strings.TrimSpace(line)
		if len(tr) == 0 {
			continue
		}
		res := svilineRe.FindStringSubmatch(line)
		if len(res) < 4 {
			continue
		}
		intfName := res[1]
		if !strings.HasPrefix(intfName, "Vl") {
			continue
		}
		vlanStr := strings.TrimPrefix(intfName, "Vl")
		vlanID, err := strconv.Atoi(vlanStr)
		if err != nil {
			continue
		}
		ret = append(ret, IntDescSVI{
			Interface:   intfName,
			VlanID:      vlanID,
			Status:      strings.ToLower(res[2]),
			Protocol:    strings.ToLower(res[3]),
			Description: strings.TrimSpace(res[4]),
		})
	}
	return ret
}
