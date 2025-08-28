package hostdata

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"slices"
	"strings"
)

type HostData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Secret   string `json:"secret"`
}

// ---
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

// ---

func GetHostAccountByGroupName(fileName string, groupName string) (result HostData, found bool) {
	// Получаем карту аккаунтов из файла, где имя_группы будет индеком.
	hd, err := getHostAccount(fileName, groupName)
	if err != nil {
		log.Println(err)
		return hd, false
	}
	return hd, true
}

// Вернуть список аккаунтов из файла конфигурации
func getHostAccount(fileName string, groupName string) (HostData, error) {

	// Данные по всем группам и соответствиям ИмяГруппы => Явки/Пароли по ней
	var hostdata HostData

	sourceFileStat, err := os.Stat(fileName)
	if err != nil {
		// Не смогли найти файл например
		log.Println(err)
		return hostdata, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		// Данный файл не является регулярным, например это директория
		log.Printf("%s is not a regular file", fileName)
		return hostdata, err
	}

	ctx := context.Background()

	// Загружаем все группы из файла.
	hdms, allGroups, err := loadAllGroupsFromFile(ctx, fileName)
	if err != nil {
		log.Printf("Failed to load groups from file %s", fileName)
		return hostdata, err
	}

	//Выполняем поиск по всем группам нужной группы
	if !slices.Contains(allGroups, groupName) {
		return hostdata, errors.New("group not found")
	}
	// Очистим его
	allGroups = make([]string, 0)
	allGroups = append(allGroups, groupName)

	// Заполняем карту данных по группам.
	hostDataInfo := GetHostDataInfo(fileName, allGroups, hdms)

	hostdata.Username = hostDataInfo.UserName
	hostdata.Password = hostDataInfo.Password
	hostdata.Secret = hostDataInfo.EnablePwd

	return hostdata, nil

}

// Функция для загрузки данных из файла groups.yaml
func loadAllGroupsFromFile(ctx context.Context, filePath string) (groupsMap map[string]map[string]string, allGroups []string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return groupsMap, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	groupsMap = make(map[string]map[string]string)

	var currentGroup string
	var connectionOptions bool

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return groupsMap, nil, ctx.Err()
		default:
			line := scanner.Text()
			if len(line) == 0 || strings.HasPrefix(strings.TrimLeft(line, " "), "#") {
				connectionOptions = false
				continue // Игнорируем пустые строки и комментарии
			}

			// Пропуск connection_options
			if strings.Contains(line, "connection_options") {
				connectionOptions = true
				continue
			}

			// Обработка новой группы
			if groupName := extractGroupName(line); groupName != "" {
				currentGroup = groupName
				allGroups = append(allGroups, groupName)
				groupsMap[currentGroup] = make(map[string]string)
				connectionOptions = false
			} else if connectionOptions {
				continue
			} else if currentGroup != "" && strings.Contains(line, "groups:") {
				// Обработка вложенных групп
				var nestedGroups []string
				for scanner.Scan() {
					nestedLine := scanner.Text()
					nestedLine = strings.TrimLeft(nestedLine, " ")
					if strings.HasPrefix(nestedLine, "#") {
						continue // Пропускаем строки с комментариями
					}
					if !strings.HasPrefix(nestedLine, "- ") {
						// Обработка ключевых параметров (username, password, secret)
						if currentGroup != "" && strings.Contains(nestedLine, "username:") {
							groupsMap[currentGroup]["username"] = extractValue(nestedLine)
						} else if currentGroup != "" && strings.Contains(nestedLine, "password:") {
							groupsMap[currentGroup]["password"] = extractValue(nestedLine)
						} else if currentGroup != "" && strings.Contains(nestedLine, "secret:") {
							groupsMap[currentGroup]["secret"] = extractValue(nestedLine)
						}
						break // Выходим из цикла при окончании списка вложенных групп
					}
					nestedGroupName := strings.TrimSpace(strings.TrimPrefix(nestedLine, "- "))
					nestedGroups = append(nestedGroups, nestedGroupName)
				}
				groupsMap[currentGroup]["groups"] = strings.Join(nestedGroups, ", ")
			} else if currentGroup != "" && (strings.Contains(line, "username:") || strings.Contains(line, "password:") || strings.Contains(line, "secret:")) {
				key := ""
				if strings.Contains(line, "username:") {
					key = "username"
				} else if strings.Contains(line, "password:") {
					key = "password"
				} else if strings.Contains(line, "secret:") {
					key = "secret"
				}
				groupsMap[currentGroup][key] = extractValue(line)
			}
		}
	}

	return groupsMap, allGroups, scanner.Err()
}

type HostIpData struct {
	Ip_e  string
	Ip_e1 string
	Ip_t1 string
	Ip_t2 string
}

func NewHostIpData(ip_e, ip_e1, ip_t1, ip_t2 string) HostIpData {

	// Реализация конструктора HostIpData
	return HostIpData{
		Ip_e:  ip_e,
		Ip_e1: ip_e1,
		Ip_t1: ip_t1,
		Ip_t2: ip_t2,
	}
}

type Response struct {
	HostName  string
	HostIp    string
	UserName  string
	Password  string
	EnablePwd string
	IpData    HostIpData
}

/*
	Example data:

	var groupsMap = make(map[string]map[string]string)
	var hostsMap = make(map[string][]string)
	var hostsIpMap = make(map[string]map[string]string)

*/

func GetHostDataInfo(hostName string, hostGroups []string, groupsMap map[string]map[string]string) Response {
	var response Response
	//var group map[string]string

	for _, groupName := range hostGroups {
		var found bool
		for !found {
			group, found := groupsMap[groupName]
			if !found {
				//fmt.Printf("Группа %s не найдена в groupsMap\n", groupName)
				break
			}

			//fmt.Printf("Информация о группе %s: %v\n", groupName, group)

			for _, key := range []string{"username", "password", "secret"} {
				if val, ok := group[key]; ok && (response.UserName == "" || response.Password == "" || response.EnablePwd == "") {
					//fmt.Printf("Добавление значения для ключа %s: %s\n", key, val)
					switch key {
					case "username":
						if response.UserName == "" {
							response.UserName = val
						}
					case "password":
						if response.Password == "" {
							response.Password = val
						}
					case "secret":
						if response.EnablePwd == "" {
							response.EnablePwd = val
						}
					}
				} //else if !ok {
				//fmt.Printf("Ключ %s не найден в группе %s\n", key, groupName)
				//}
			}

			parentGroups := extractGroupsFromGroup(groupName, groupsMap)
			//if parentGroups != nil {
			if len(parentGroups) == 0 || (response.UserName != "" && response.Password != "" && response.EnablePwd != "") {
				break
			}
			groupName = parentGroups[0]
			//}
		}
	}

	return response

}

/*
func GetHostInfo(hostName string, hostsMap map[string][]string, hostsIpMap map[string]map[string]string, groupsMap map[string]map[string]string) Response {
	var response Response

	hostGroups, ok := hostsMap[hostName]
	if !ok {
		fmt.Println("Хост не найден")
		return response
	}

	hostIpData, ok := hostsIpMap[hostName]
	if !ok {
		fmt.Println("IP хоста не найдено")
	} else {
		// Заполняем ответ дополнительными IP.
		hipd := NewHostIpData(hostIpData["hoost_e"], hostIpData["hoost_e2"], hostIpData["hoost_t1"], hostIpData["hoost_t2"])
		response.IpData = hipd
	}

	response.HostName = hostName
	hostIp, ok := hostsIpMap[hostName]["hostname"]
	if !ok {
		fmt.Println("IP хоста не найдено")
	} else {
		response.HostIp = hostIp
	}
	fmt.Printf("Группы для хоста %s: %v\n", hostName, hostGroups)

	for _, groupName := range hostGroups {
		var found bool
		for !found {
			group, found := groupsMap[groupName]
			if !found {
				fmt.Printf("Группа %s не найдена в groupsMap\n", groupName)
				break
			}

			fmt.Printf("Информация о группе %s: %v\n", groupName, group)

			for _, key := range []string{"username", "password", "secret"} {
				if val, ok := group[key]; ok && (response.UserName == "" || response.Password == "" || response.EnablePwd == "") {
					fmt.Printf("Добавление значения для ключа %s: %s\n", key, val)
					switch key {
					case "username":
						if response.UserName == "" {
							response.UserName = val
						}
					case "password":
						if response.Password == "" {
							response.Password = val
						}
					case "secret":
						if response.EnablePwd == "" {
							response.EnablePwd = val
						}
					}
				} else if !ok {
					fmt.Printf("Ключ %s не найден в группе %s\n", key, groupName)
				}
			}

			parentGroups := extractGroupsFromGroup(groupName, groupsMap)
			if parentGroups != nil {
				if len(parentGroups) == 0 || (response.UserName != "" && response.Password != "" && response.EnablePwd != "") {
					break
				}

				groupName = parentGroups[0]
			}
		}
	}

	return response
}
*/

func extractGroupsFromGroup(groupName string, groupsMap map[string]map[string]string) []string {
	var groups []string
	group, ok := groupsMap[groupName]
	if ok {
		groupsStr, ok := group["groups"]
		if ok {
			// Разбиваем строку на срез строк, если это необходимо
			// Например, если groupsStr содержит список групп, разделенных запятыми
			groups = strings.Split(groupsStr, ",")
		}
	}
	return groups
}

/*
func extractGroupsFromGroup(groupName string, groupsMap map[string]map[string]string) []string {
	var groups []string

	// Проверяем наличие группы в maps
	group, ok := groupsMap[groupName]
	if !ok {
		return nil // Возвращаем nil вместо пустого slices, если группы отсутствуют
	}

	// Проверяем наличие ключа "groups"
	groupsStr, ok := group["groups"]
	if !ok {
		return nil // Возвращаем nil, если ключ "groups" отсутствует
	}

	// Разбиваем строку на срез, если это необходимо
	groups = strings.Split(groupsStr, ",")
	return groups
}
*/
