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

	hd, err := GetHostAccount(fileName)
	if err != nil {
		log.Println(err)
	}

	if out, ok := hd[groupName]; ok {
		return out, true
	} else {
		return HostData{}, false
	}
}

func GetHostAccount(fileName string) (map[string]HostData, error) {

	//var hostDataMap map[string]HostData // Данные по всем группам и соответствиям ИмяГруппы => Явки/Пароли по ней

	// Данные по всем группам и соответствиям ИмяГруппы => Явки/Пароли по ней
	hostDataMap := make(map[string]HostData) // Данные по всем группам и соответствиям ИмяГруппы => Явки/Пароли по ней

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
		//out = append(out, hd)
	}

	return hostDataMap, nil

}
