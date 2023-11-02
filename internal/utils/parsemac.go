package utils

import (
	"regexp"
)

type MacLineData struct {
	vlan  string
	mac   string
	iface string
}

type HostMacLineData struct {
	HostName string
	mld      []MacLineData
}

func NewHostMacLineData(hostname string) *HostMacLineData {
	return &HostMacLineData{
		HostName: hostname,
		mld:      []MacLineData{},
	}
}

func NewMacLineData(
	vlan string,
	mac string,
	iface string,
) *MacLineData {

	return &MacLineData{
		vlan:  vlan,
		mac:   mac,
		iface: iface,
	}
}

func ParseMacLine(line string) MacLineData {

	/*
	   1    548a.ba01.50b3    DYNAMIC     Gi0/43
	   1    b022.7a2e.5561    DYNAMIC     Gi0/43
	   19    805e.c02d.4d50    DYNAMIC     Gi0/43
	   204    0000.aa8d.ada8    DYNAMIC     Gi0/43
	   "   1      0e55.6c89.a819   dynamic ip,ipx,assigned,other Port-channel20             "
	*/

	re, _ := regexp.Compile(`^\s*(\d+)\s+(\S+)\s+([D|S]\S+)\s{4,6}(\S+)`)
	res := re.FindStringSubmatch(line)

	if len(res) > 0 {
		return *NewMacLineData(res[1], res[2], res[4])
	}
	re, _ = regexp.Compile(`^\s*(\d+)\s+(\S+)\s+([D|S|d]\S+)\s+ip\S+\s+(\S+)`)
	res = re.FindStringSubmatch(line)
	if len(res) > 0 {
		return *NewMacLineData(res[1], res[2], res[4])
	}

	return MacLineData{}

}

func (m *MacLineData) GetVlan() string {
	return m.vlan
}

func (m *MacLineData) GetMac() string {
	return m.mac
}

func (m *MacLineData) GetIface() string {
	return m.iface
}

func (m *MacLineData) GetLen() int {
	return len(m.vlan) + len(m.mac) + len(m.iface)
}

/*


// IntedStringToInts - сконвертировать массив string (в которых только числа) в массив Integer
func IntedStringToInts(strarr []string) []int {
	var out []int
	for _, v := range strarr {
		nmbr, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		out = append(out, nmbr)
	}
	return out
}

// ParseMacFile - открыть файл с MAC-адресами, и распасрсить его в массив хостов с данными.
func ParseMacFile(macFileName string) ([]HostMacLineData, error) {

	fmt.Println("Parse MAC file:", macFileName)

	MacLines := []MacLineData{}  // Временный массив
	var output []HostMacLineData // Исходящие данныы

	// Читаем ACL файл
	aclFile, err := os.OpenFile(macFileName, os.O_RDONLY, 0644)
	if err != nil {
		return output, fmt.Errorf("ошибка открытия файла: %s", err)
	}
	defer aclFile.Close()

	scanner := bufio.NewScanner(aclFile)
	scanner.Split(bufio.ScanLines)

	// Строки ACL файла
	var aclFileLines []string

	for scanner.Scan() {
		aclFileLines = append(aclFileLines, scanner.Text())
	}
	aclFile.Close()

	var hostName string
	var hmld HostMacLineData
	for _, s := range aclFileLines {
		tr := strings.TrimSpace(s)
		if len(tr) > 0 {

			if strings.Contains(tr, "hostgetmac:") {
				hostName = strings.TrimPrefix(tr, "hostgetmac: ")
				if len(MacLines) > 0 {
					hmld.mld = MacLines
					output = append(output, hmld)
				}
				// Новая
				hmld = *NewHostMacLineData(hostName)

			} else {
				a := parseArpLine(tr)
				if len(a.vlan) > 0 {
					MacLines = append(MacLines, a)
				}
			}
		}
	}
	// Добавляем последний проверяемый в выходной массив
	if len(MacLines) > 0 {
		hmld.mld = MacLines
		output = append(output, hmld)
	}

	return output, nil

}
*/
