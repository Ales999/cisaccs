package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type VlanLineData struct {
	vlid   int
	vlname string
}

func NewVlanLineData(vlid int, vlname string) *VlanLineData {
	return &VlanLineData{
		vlid:   vlid,
		vlname: vlname,
	}
}

func (a *VlanLineData) GetId() int {
	return a.vlid
}

func (a *VlanLineData) GetName() string {
	return a.vlname
}

func (a *VlanLineData) PrintData() {
	/*
		fmt.Println("ID: ", a.vlid)
		fmt.Println("VlanName: ", a.vlname)
	*/
	if a.vlid > 0 {
		fmt.Printf("ID %d, Name: %s\n", a.vlid, a.vlname)
	}
}

// Распарсить одну строку
func ParseVlan(line string) VlanLineData {
	/*
	   1    default                          active
	   141  TD_*14.16/28                     active
	   244  ESN                              active
	   1002 fddi-default                     act/unsup
	   1003 trcrf-default                    act/unsup
	   1004 fddinet-default                  act/unsup
	   1005 trbrf-default                    act/unsup
	   2001 VLAN2001                         active
	*/
	tr := strings.TrimSpace(line)
	if len(tr) > 2 {

		// Парсим строку что cisco выдала
		re, _ := regexp.Compile(`^(\d{1,4})\s+(\S+)\s+(\S+)\s*`)
		res := re.FindStringSubmatch(tr)
		// Если что-то в матчинге есть то согдает запись и возвращаем ее
		if len(res) > 0 {
			vlid, err := strconv.Atoi(res[1])
			if err != nil {
				// ... handle error
				return VlanLineData{}

			}
			return *NewVlanLineData(vlid, res[2]) // *NewArpLineData(res[1], res[2], res[3])
		}
	}

	return VlanLineData{}

}
