package crawlerprovider

import (
	. "cryptopackage/model"
	"encoding/json"
	"fmt"

	. "huobicagent/model"
	"tczbgo/config"
	"tczbgo/kafka"
	"tczbgo/logger"
	"tczbgo/system/zbhttp"
	"tczbgo/system/zbos"
)

var (
	appSettings            AppSettings
	getPageApiUrl          string
	heartbeatApiUrl        string
	sendStopApiUrl         string
	machineHeartBeatApiUrl string
	provider               string
	heartBeat              string
	version                string
	topic                  string
)

func init() {
	provider, _ = zbos.HostName()
	version = config.Version
	topic = kafka.Topic
	config.GetAppSettings(&appSettings)
	getPageApiUrl = fmt.Sprintf("%v%v/%v", appSettings.GetPageApiUrl, "crypto", provider)
	heartbeatApiUrl = fmt.Sprintf("%v%v", appSettings.HeartBeatApiUrl, provider)
	sendStopApiUrl = fmt.Sprintf("%v%v", appSettings.SendStopApiUrl, provider)
	machineHeartBeatApiUrl = fmt.Sprintf(appSettings.MachineHeartBeatApiUrl, provider)
}

func CheckPage(url string) int {
	statusCode, _, err := zbhttp.NewHttp(zbhttp.Method.Head, url, nil, nil)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("checkpage fail, Error: %v", err.Error()))
		return 0
	}
	return statusCode
}

func GetPage() Page {
	var page Page
	_, response, err := zbhttp.Get(getPageApiUrl)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("getpage get fail, Error: %v", err.Error()))
		return page
	} else {
		err = json.Unmarshal(response, &page)
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("getpage unmarshal fail, Error: %v", err.Error()))
		}
	}
	return page
}

func SendStop(pageName string) {
	api := fmt.Sprintf("%v/%v", sendStopApiUrl, pageName)

	_, _, err := zbhttp.NewHttp(zbhttp.Method.Patch, api, nil, nil)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("sendstop patch fail, Error: %v", err.Error()))
	}
}

func SendHeartbeat(pageList []string) {
	pageData, err := json.Marshal(pageList)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("heartbeat marshal fail, Error: %v", err.Error()))
		return
	}
	header := map[string][]string{"Content-Type": {"application/json"}}
	_, _, err = zbhttp.NewHttp(zbhttp.Method.Patch, heartbeatApiUrl, pageData, header)

	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("heartbeat patch fail, Error: %v", err.Error()))
	}
}

func SendMachineHeartBeat(pageList []string) {
	heartBeat = fmt.Sprintf("W:%v,V:%v,H:", len(pageList), version)
	statusData, err := json.Marshal(heartBeat)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("machineheartbeat marshal fail, Error: %v", err.Error()))
		return
	}
	_, _, err = zbhttp.Post(fmt.Sprintf("%v?status=%v", machineHeartBeatApiUrl, heartBeat), statusData)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("machineheartbeat patch fail, Error: %v", err.Error()))
	}
}

func SendData(cryptoData []CryptoData) {
	cryptoDataByte, err := json.Marshal(cryptoData)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("send kafka marshal fail, Error: %v", err.Error()))
	} else {
		err = kafka.SendMessage(topic, cryptoDataByte)
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("send kafka marshal fail, Error: %v", err.Error()))
		}
	}
}
