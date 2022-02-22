package crawlerservice

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"tczbgo/config"
	"tczbgo/logger"
	"tczbgo/system/zbtime"
	"time"

	"huobicagent/domainservice/crawlertransfer"
	"huobicagent/infrastructure/crawlerprovider"
	. "huobicagent/model"

	"github.com/go-rod/rod"
)

var (
	appSettings AppSettings
	maxWork     int
	htmlChan    chan string = make(chan string)
	pageList                = []string{}
	browser     *rod.Browser
)

func init() {
	config.GetAppSettings(&appSettings)
	maxWork = appSettings.MaxWork
}

func HuobiChromedp() {
	now := time.Now().Format("15:04:05")
	logger.Log(logger.LogLevel.Debug, fmt.Sprintf("agent start: %v", now))
	go setTimeCloseExe()
	go heartBeat()
	go processedHtml()
	go setupCloseHandler()
	go focusPage()

	t := time.NewTicker(60 * time.Second)
	for {
		<-t.C
		if len(pageList) < maxWork {
			go getPage()
		}
	}
}

func setTimeCloseExe() {
	t := 6 * time.Hour
	min := -30
	max := 30
	rangeTime := rangeRadom(min, max)
	t = t + time.Duration(rangeTime)*time.Minute
	time.AfterFunc(t, closeExe)
}

func rangeRadom(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func closeExe() {
	now := time.Now().Format("15:04:05")
	logger.Log(logger.LogLevel.Debug, fmt.Sprintf("agent close: %v", now))
	closeBrowser()
	time.Sleep(5 * time.Second)
	os.Exit(0)
}

func processedHtml() {
	defer panicRecover("service/processedHtml")
	count := 0
	for {
		htmlContent := <-htmlChan
		data, dataError := crawlertransfer.GetHuobiData(htmlContent, zbtime.UnixTimeNow(zbtime.Duration.Millisecond))
		if !dataError {
			crawlerprovider.SendData(data)
			fmt.Printf("%v, %v/%v\n", time.Now().Format("15:04:05"), data[0].Datum, data[0].Name)
		}

		if count == 100 {
			count = 0
		} else {
			count++
		}
	}
}

func heartBeat() {
	defer panicRecover("service/heartBeat")
	t := time.NewTicker(60 * time.Second)
	for {
		<-t.C
		fmt.Println(fmt.Sprintf("%v", pageList))
		crawlerprovider.SendHeartbeat(pageList)
		crawlerprovider.SendMachineHeartBeat(pageList)
	}

}

func getPage() {
	defer panicRecover("service/getPage")

	if browser == nil {
		browser = rod.New().MustConnect()
	}
	var page Page
	page = crawlerprovider.GetPage()
	if (page != Page{}) {
		errCount := checkPage(page.Url, page.PageName, 3)
		if errCount < 3 {
			pageList = append(pageList, page.PageName)
			openPage(page.Url, page.PageName)
		} else {
			crawlerprovider.SendStop(page.PageName)
		}
	}
}

func focusPage() {
	defer panicRecover("service/focusPage")
	t := time.NewTicker(10 * time.Minute)
	for {
		<-t.C
		if browser != nil {
			pages, err := browser.Pages()
			if err != nil {
				logger.Log(logger.LogLevel.Error, fmt.Sprintf("get browserpage fail, Error: %v", err.Error()))
			} else if len(pages) > 0 {
				for _, page := range pages {
					_, err = page.Activate()
					if err != nil {
						logger.Log(logger.LogLevel.Error, fmt.Sprintf("page focus fail, Error: %v", err.Error()))
					}
					time.Sleep(2 * time.Second)
				}
			}
		}
	}
}

func sendStop(pageName string) {
	crawlerprovider.SendStop(pageName)
	for index, page := range pageList {
		if page == pageName {
			pageList = stringSliceRemove(pageList, index)
			break
		}
	}
}

func stringSliceRemove(slice []string, index int) []string {
	next := index + 1
	return append(slice[:index], slice[next:]...)
}

func openPage(url string, pageName string) {
	timeOut := false
	count := 0
	page := browser.MustPage(url)

	time.AfterFunc(6*time.Hour, func() {
		timeOut = true
		sendStop(pageName)
		page.Close()
	})

	t := time.NewTicker(1 * time.Second)
	for !timeOut {
		<-t.C
		el := page.MustElement(".page-global-exchange")
		if el != nil {
			htmlChan <- el.MustHTML()
		}
		if count >= 1800 {
			page.Reload()
			count = 0
		}
	}
}

func checkPage(url string, pageName string, checkCount int) int {
	errCount := 0
	status := crawlerprovider.CheckPage(url)
	pass := status == 200
	for !pass && errCount < checkCount {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("checkpahe, Error: %v, %v", url, status))
		errCount++
		if errCount > 1 {
			errorData := crawlertransfer.GetErrorData(pageName, zbtime.UnixTimeNow(zbtime.Duration.Millisecond))
			crawlerprovider.SendData(errorData)
		}
		time.Sleep(60 * time.Second)
		status = crawlerprovider.CheckPage(url)
		pass = status == 200
	}

	return errCount
}

func closeBrowser() {
	for _, pageName := range pageList {
		go crawlerprovider.SendStop(pageName)
	}
}

func setupCloseHandler() {
	defer panicRecover("service/setupCloseHandler")
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal, wait Close Browse & Send stop, wait 5 second")
		closeExe()
	}()
}

func panicRecover(funcName string) {
	if r := recover(); r != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("%v panicRecover, Error: %v", funcName, r))
	}
}
