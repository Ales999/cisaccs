package utils

import (
	"regexp"
)

type MacLineData struct {
	vlan  string
	mac   string
	iface string
}

// macReWithProto ловит старый формат Cat4500: "1    mac    dynamic ip,ipx,... Port-channel20"
// macReSimple ловит новый формат (Cat9k/9500/9410): "594    mac    DYNAMIC     Po49"
// Ловим и DYNAMIC, и STATIC (case-insensitive). На C9200L хосты часто сидят на
// access-портах как STATIC — их нужно отслеживать. А вот STATIC на интерфейсе
// VLAN (VlXXX) это SVI-gateway (системный MAC), он отфильтрован в ParseMacLine.
var (
	macReWithProto = regexp.MustCompile(`(?i)^\s*(\d+)\s+(\S+)\s+(dynamic|static)\s+ip\S+\s+(\S+)`)
	macReSimple    = regexp.MustCompile(`(?i)^\s*(\d+)\s+(\S+)\s+(dynamic|static)\s+(\S+)`)
)

// reVlanIface ловит интерфейсы SVI/ВLAN в выводе mac-table («Vl40», «Vlan40»).
// Это шлюзы VLAN (системные MAC), а не конечные хосты, поэтому они не отслеживаются.
var reVlanIface = regexp.MustCompile(`(?i)^vl(an)?\d`)

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
	   Old format (Cat4500):
	   1      0e55.6c89.a819   dynamic ip,ipx,assigned,other Port-channel20
	   3161      0008.e3ff.fd90    static ip,ipx,assigned,other Switch

	   New format (Cat9k/9500/9410):
	   594    f433.9234.d502    STATIC      Vl594
	   583    6c0b.5e0d.a348    DYNAMIC     Po49
	*/

	var mld MacLineData
	if res := macReWithProto.FindStringSubmatch(line); len(res) >= 5 {
		mld = *NewMacLineData(res[1], res[2], res[4])
	} else if res := macReSimple.FindStringSubmatch(line); len(res) >= 5 {
		mld = *NewMacLineData(res[1], res[2], res[4])
	}

	// MAC на интерфейсе VLAN (SVI-gateway) — не конечная точка, не отслеживаем.
	if mld.GetLen() > 0 && reVlanIface.MatchString(mld.GetIface()) {
		return MacLineData{}
	}

	return mld

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
