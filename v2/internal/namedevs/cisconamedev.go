package namedevs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// CiscoNameDev содержит в себе имя устройства (cisco) и его параметры для подключения
//
// PS: Кроме данных с именами и паролем
type CiscoNameDev struct {
	NameDev   string // Имя узла (hostname)
	Group     string // Имя Группы.
	HostIp    string // Ip данного узла (management).
	HostExtIp string // Внешний IP
	Iface     string // Имя интерфейса которым подключен у вышестояшего коммутатора, если есть.
}

// newCiscoNameDev  - вернуть ссылку на новый экземпляр структуры
func newCiscoNameDev(namedev string, group string, hostip string, hostextip string, iface string) *CiscoNameDev {
	return &CiscoNameDev{
		NameDev:   namedev,
		Group:     group,
		HostIp:    hostip,
		HostExtIp: hostextip,
		Iface:     iface,
	}
}

type CiscoNameDevs CiscoNameDev

func (c *CiscoNameDevs) GetHostDataByHostName(cisFileName string, hostName string) (*CiscoNameDev, error) {

	ctx := context.Background()
	// Load hosts from config file.
	hostsMap, hostsIpMap, err := loadHosts(ctx, cisFileName)
	if err != nil {
		return nil, errors.New("Ошибка при чтении/парсинге файла groups.yaml" + err.Error() + " " + cisFileName)
	}
	//info := GetHostInfo(hostName)

	hostGroups, ok := hostsMap[hostName]
	if !ok {
		//fmt.Println("Хост не найден")
		return nil, errors.New("хост не найден в группе")
	}

	hostIpData, ok := hostsIpMap[hostName]
	if !ok {
		//fmt.Println("IP хоста не найдено")
		return nil, errors.New("IP хоста не найдено")
	} //else {
	// Заполняем ответ дополнительными IP.
	//hipd := NewHostIpData(hostIpData["hoost_e"], hostIpData["hoost_e2"], hostIpData["hoost_t1"], hostIpData["hoost_t2"])
	//response.IpData = hipd
	//}

	//response.HostName = hostName
	hostIp, ok := hostsIpMap[hostName]["hostname"]
	if !ok {
		//fmt.Println("IP хоста не найдено")
		return nil, errors.New("IP хоста не найдено")
	} //else {
	//response.HostIp = hostIp
	//}

	// test debug output
	//fmt.Printf("Группы для хоста %s: %v\n", hostName, hostGroups)
	if len(hostGroups) <= 0 {
		return nil, errors.New("группа для хоcта не найдена")
	}

	ret := newCiscoNameDev(hostName, hostGroups[0], hostIp, hostIpData["hoost_e"], "-none-")

	//fmt.Println(ret)
	return ret, nil
	//return newCiscoNameDev(hostName, hostGroups[0], hostIp, hostIpData["hoost_e"], "-none-"), nil

	/*
			//dev := hostName
			var cnd = newCiscoNameDev(
				strings.ToLower(hostName),
				kcis.String(hostName+".group"),
				kcis.String(hostName+".host"),
				kcis.String(hostName+".hoost_e"),
				kcis.String(hostName+".spb4face"),
			)
		//return cnd, nil
	*/

}

// findHostsInGroup - Получает список хостов, входящих в указанную группу.
//
// Параметры:
//
//	hostsByGroup: карта памяти, где ключ - имя хоста, а значение - массив имен групп.
//	groupName: имя группы, для поиска хостов.
//
// Возвращает:
//
//	массив строк, содержащий имена хостов, входящих в указанную группу.
//	Если группа не найдена, возвращает пустой массив.
func findHostsInGroup(hostsByGroup map[string][]string, groupName string) []string {
	var hosts []string

	for host, groups := range hostsByGroup {
		for _, g := range groups {
			if g == groupName {
				hosts = append(hosts, host)
				break // Чтобы не добавлять хост несколько раз, если он входит в группу несколько раз
			}
		}
	}

	return hosts
}

// GetHostsByGroupName - получить список хостов относящийся к заданной группе
func (c *CiscoNameDevs) GetHostsByGroupName(cisFileName string, grpName string) ([]string, error) {

	ctx := context.Background()
	// Load hosts from config file.
	hostsMap, _, err := loadHosts(ctx, cisFileName)
	if err != nil {
		return nil, errors.New("Ошибка при чтении/парсинге файла groups.yaml" + err.Error() + " " + cisFileName)
	}
	/*
		// Получаем список всех ключей из карты - получим сртслк самих хостов.
		allHosts := make([]string, len(hostsMap))
		i := 0
		for k := range hostsMap {
			allHosts[i] = k
			i++
		}
	*/
	// Получаем список всех ключей из карты
	hosts := findHostsInGroup(hostsMap, grpName)
	if len(hosts) == 0 {
		return nil, errors.New("группа не найдена в файле groups.yaml")
	}

	return hosts, nil
}

// --- V2 ---

// карта хостов
//var hostsMap = make(map[string][]string)
//var hostsIpMap = make(map[string]map[string]string)

func loadHosts(ctx context.Context, filePath string) (hostsMap map[string][]string, hostsIpMap map[string]map[string]string, err error) {

	hostsMap = make(map[string][]string)
	hostsIpMap = make(map[string]map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var txtlines []string
	// Загрузим весь файл в массив строк
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
			_str := scanner.Text()
			// Пропустим комментарии
			if strings.HasPrefix(strings.TrimLeft(_str, " "), "#") {
				continue
			}
			txtlines = append(txtlines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("ошибка чтения файла: %s", err)
	}

	var currentHost string

	for n, line := range txtlines {
		// ---------------------
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
			// Пропускаем пустые строки и комментарии - ищем начала блока имени хоста
			if len(line) == 0 || strings.HasPrefix(strings.TrimLeft(line, " "), "#") {
				continue
			}
			// Ищем новый хост
			if !strings.HasPrefix(line, " ") && strings.Contains(line, ":") { // && !(strings.Contains(line, "groups:") || strings.Contains(line, "hostname:")) {
				// Начинаем новый хост
				if strings.EqualFold(strings.TrimLeft(line, " "), line) {
					currentHost = extractGroupName(line)
					hostsMap[currentHost] = []string{}
					hostsIpMap[currentHost] = make(map[string]string)
					hostsIpMap[currentHost]["hostname"] = ""
					hostsIpMap[currentHost]["hoost_e"] = ""
					hostsIpMap[currentHost]["hoost_e2"] = ""
					hostsIpMap[currentHost]["hoost_t1"] = ""
					hostsIpMap[currentHost]["hoost_t2"] = ""

					// Читаем блок данных за этом хостом

					// Выбираем остатки что еще не сканировали в отдельный слайс (только следующие 20 строк)
					var tlsts []string
					if len(txtlines[n+1:]) > 22 {
						tlsts = txtlines[n+1 : n+20]
					} else {
						tlsts = txtlines[n+1:]
					}
					for f, tlst := range tlsts {

						// Иначе ищем данные дальше.
						if strings.Contains(tlst, "hostname:") {
							hostIp := extractValue(tlst)
							hostsIpMap[currentHost]["hostname"] = hostIp
						}
						// Блок Группы хоста
						if strings.Contains(tlst, "groups:") {
							// Debug print
							//fmt.Println("Groups found")
							var groups []string
							//var groupLines []string
							grpsts := tlsts[f+1:]
							for _, grp := range grpsts {
								// Удалим профекс из пробелов
								grp = strings.TrimLeft(grp, " ")
								if strings.HasPrefix(grp, "- ") {
									groups = append(groups, strings.TrimSpace(grp[2:]))
								} else if len(grp) == 0 {
									break
								}
							}
							hostsMap[currentHost] = groups
						}
						// Блок - Дополнительные IP адреса
						if strings.Contains(tlst, "data:") {
							//fmt.Println("Data found")
							// Дополнительные данные хоста
							dataLines := tlsts[f+1:]
							for _, dataLine := range dataLines {
								// Если это не комментарий и не пустая строка
								if len(dataLine) > 0 && !strings.HasPrefix(strings.TrimLeft(dataLine, " "), "#") {
									if strings.HasPrefix(strings.TrimLeft(dataLine, " "), "hoost") {
										_line := strings.TrimLeft(dataLine, " ")
										if strings.Contains(_line, "hoost_e") {
											hostsIpMap[currentHost]["hoost_e"] = extractValue(_line)
										} else if strings.Contains(_line, "hoost_e2") {
											hostsIpMap[currentHost]["hoost_e2"] = extractValue(_line)
										} else if strings.Contains(_line, "hoost_t1") {
											hostsIpMap[currentHost]["hoost_t1"] = extractValue(_line)
										} else if strings.Contains(_line, "hoost_t2") {
											hostsIpMap[currentHost]["hoost_t2"] = extractValue(_line)
										}
									}
								}
								// Если строка пустая, то прерываем текущий вложенный цикл.
								if len(dataLine) == 0 {
									break
								}
							}
						}

						// Возьмем следующую строку и проверим если она не пустая и начинается с пробела
						if len(tlsts) >= f+1 { // Проверка выходя за пределы массива
							if len(tlsts[f+1:]) > 0 {
								nextLine := tlsts[f+1:][0]
								// Если блок хоста заканчивается, то прерываем данный цикл - следующий блок будет считаться новым хостом
								if len(nextLine) > 0 && !strings.HasPrefix(nextLine, " ") {
									break
								}
							}
						}
					}

				}
			}
		}
	}

	return hostsMap, hostsIpMap, scanner.Err()
}

// extractValue - функция для извлечения значения из строки в формате "ключ:значение"
func extractValue(line string) (value string) {
	// Разделим строку на две части: ключ и значение (последний двоеточий)
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return ""
	}

	// Получим значение после последнего двоеточия
	value = strings.TrimSpace(parts[1])

	// Если значение пустое, вернем пустую строку
	if value == "" {
		return ""
	}

	// Разделим значение по пробелам и возьмем первое слово
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	value = fields[0]

	// Уберем кавычки с начала и конца строки, если они есть
	value = removeQuotes(value)

	return value
}

// extractGroupName - функция для парсинга имени группы из строки.
func extractGroupName(line string) string {
	if strings.TrimLeft(line, " ") == line {
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return ""
}

// Функция для удаления кавычек с начала и конца строки
func removeQuotes(s string) string {
	if len(s) < 2 {
		return s
	}

	quotes := []string{"\"", "'", "`"}
	for _, quote := range quotes {
		if s[0] == quote[0] && s[len(s)-1] == quote[0] {
			return s[1 : len(s)-1]
		}
	}
	return s
}
