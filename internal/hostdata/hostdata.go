package hostdata

import (
	"encoding/json"
	"log"
	"os"
)

type HostData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Secret   string `json:"secret"`
}

func GetHostAccountByGroupName(fileName string, groupName string) (result HostData, found bool) {
	// Получаем карту аккаунтов из файла, где имя_группы будет индеком.
	hd, err := GetHostAccount(fileName)
	if err != nil {
		log.Println(err)
	}
	// Поиск в карте нужного по имени группы
	if out, ok := hd[groupName]; ok {
		return out, true
	} else {
		return HostData{}, false
	}
}

// Вернуть список аккаунтов из файла конфигурации
func GetHostAccount(fileName string) (map[string]HostData, error) {

	// Данные по всем группам и соответствиям ИмяГруппы => Явки/Пароли по ней
	hostDataMap := make(map[string]HostData)

	sourceFileStat, err := os.Stat(fileName)
	if err != nil {
		// Не смогли найти файл например
		log.Println(err)
		return nil, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		// Данный файл не является регулярным, например это директория
		log.Printf("%s is not a regular file", fileName)
		return nil, err
	}

	// Открываем файл как byte[]
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var objmap map[string]json.RawMessage

	err = json.Unmarshal(fileData, &objmap)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for key := range objmap {
		var hd HostData
		err = json.Unmarshal(objmap[key], &hd)
		if err != nil {
			return nil, err
		}
		hostDataMap[key] = hd
	}

	return hostDataMap, nil

}
