package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type ArpLineData struct {
	ip    string
	mac   string
	iface string
}

func NewArpLineData(ip string, mac string, iface string) *ArpLineData {
	return &ArpLineData{
		ip:    ip,
		mac:   mac,
		iface: iface,
	}

}

func (a *ArpLineData) GetIp() string {
	return a.ip
}

func (a *ArpLineData) GetMac() string {
	return a.mac
}

func (a *ArpLineData) GetIface() string {
	return a.iface
}

func (a *ArpLineData) GetLen() int {
	return len(a.iface) + len(a.ip) + len(a.mac)
}

func (a *ArpLineData) PrintData() {
	fmt.Printf("IP: %v\n", a.ip)
	fmt.Printf("MAC: %v\n", a.mac)
	fmt.Printf("IFaces: %v\n", a.iface)

}

// ParseArpLine - парсим строку вывода cisco 'sh arp'
func ParseArpLine(line string) ArpLineData {

	/*
	   Protocol  Address          Age (min)  Hardware Addr   Type   Interface
	   Internet  10.23.1.1             172   aabb.cc00.1030  ARPA   Ethernet0/3
	   Internet  10.23.1.2               -   aabb.cc00.2030  ARPA   Ethernet0/3
	*/

	//fmt.Println(line)
	tr := strings.TrimSpace(line)

	re, _ := regexp.Compile(`^Internet  (\S+)\s+[0-9|-]+\s+(\S+)\s+ARPA\s+(\S+)$`)
	res := re.FindStringSubmatch(tr)

	if len(res) > 0 {
		if res[2] != "Incomplete" {
			return *NewArpLineData(res[1], res[2], res[3])
		}
	}

	return ArpLineData{}

}

/*
func ParseARPFile(arplFileName string) ([]ArpLineData, error) {

	out := []ArpLineData{}

	// Читаем ACL файл
	arpFile, err := os.OpenFile(arplFileName, os.O_RDONLY, 0644)
	if err != nil {
		return out, fmt.Errorf("ошибка открытия файла: %s", err)
	}
	defer arpFile.Close()

	scanner := bufio.NewScanner(arpFile)
	scanner.Split(bufio.ScanLines)

	// Строки ACL файла
	var arpFileLines []string

	for scanner.Scan() {
		arpFileLines = append(arpFileLines, scanner.Text())
	}
	arpFile.Close()

	for _, s := range arpFileLines {
		tr := strings.TrimSpace(s)
		if len(tr) > 0 {
			out = append(out, ParseArpLine(tr))
		}
	}

	return out, nil

}
*/
