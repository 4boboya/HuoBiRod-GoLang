package crawlertransfer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"tczbgo/config"
	"tczbgo/logger"
	"tczbgo/system/zbtime"

	. "cryptopackage/model"
	. "huobicagent/model"

	"github.com/PuerkitoBio/goquery"
)

var (
	appSettings      AppSettings
	site             string
	cryptoFindFloat  = make(map[string]string)
	cryptoFindString = make(map[string]string)
	cryptoReg        = make(map[string]*regexp.Regexp)
	getErrorDatumReg = regexp.MustCompile(`.+_`)
	getErrorNameReg  = regexp.MustCompile(`_.+`)
)

func init() {
	config.GetAppSettings(&appSettings)
	site = appSettings.Site
	cryptoFindFloat["price"] = ".price-container > .price"
	cryptoFindFloat["volume"] = "dl[class=amount] > dd"
	cryptoFindString["name"] = ".nuxt-link-exact-active > .name"
	cryptoFindString["datum"] = ".group > .active"
	cryptoReg["price"] = regexp.MustCompile(``)
	cryptoReg["volume"] = regexp.MustCompile(`,|\s\S+`)
}

func GetErrorData(pageName string, RequestTime int64) []CryptoData {
	var cryptoDataList []CryptoData
	var cryptoData CryptoData

	cryptoData.Site = site
	cryptoData.RequestTime = RequestTime
	cryptoData.Type = Type.Crypto
	cryptoData.Name = getErrorNameReg.ReplaceAllString(pageName, "")
	cryptoData.Datum = getErrorDatumReg.ReplaceAllString(pageName, "")
	cryptoData.CryptoBasic.Price = Empty.Price
	cryptoData.CryptoBasic.Volume = Empty.Volume

	cryptoDataList = append(cryptoDataList, cryptoData)

	return cryptoDataList
}

func GetHuobiData(htmlContent string, RequestTime int64) ([]CryptoData, bool) {
	var dataError bool = false
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))

	var cryptoDataList []CryptoData
	var cryptoData CryptoData

	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("NewDocumentFromReader fail, Error: %v", err.Error()))
		dataError = true
		return cryptoDataList, dataError
	}

	doc.Find("div[class=ticker]").Each(func(_ int, titleSelection *goquery.Selection) {
		var dataErrors []bool = []bool{false, false, false, false}
		titleString, err := goquery.OuterHtml(titleSelection)
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("OuterHtml fail, Error: %v", err.Error()))
			return
		}
		titleData, err := goquery.NewDocumentFromReader(strings.NewReader(titleString))
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("getTitle NewDocumentFromReader fail, Error: %v", err.Error()))
			return
		}
		cryptoData.Site = site
		cryptoData.RequestTime = RequestTime
		cryptoData.Type = Type.Crypto
		cryptoData.Name, dataErrors[0] = getDataString(doc, "name")
		cryptoData.Datum, dataErrors[1] = getDataString(doc, "datum")
		cryptoData.CryptoBasic.Price, dataErrors[2] = getDataFloat(titleData, "price")
		cryptoData.CryptoBasic.Volume, dataErrors[3] = getDataFloat(titleData, "volume")

		for _, error := range dataErrors {
			if error {
				dataError = true
			}
		}
		cryptoData.SendTime = zbtime.UnixTimeNow(zbtime.Duration.Millisecond)
		cryptoDataList = append(cryptoDataList, cryptoData)
		cryptoData = CryptoData{}
	})
	return cryptoDataList, dataError
}

func getDataString(doc *goquery.Document, dataType string) (string, bool) {
	var data string
	dataError := false
	replaceWhite := regexp.MustCompile(`\s`)
	doc.Find(cryptoFindString[dataType]).Each(func(_ int, selection *goquery.Selection) {
		dataString, err := goquery.OuterHtml(selection)
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("get data string outerhtml fail, Error: %v", err.Error()))
			dataError = true
		} else if dataDoc, err := goquery.NewDocumentFromReader(strings.NewReader(dataString)); err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("get data string NewDocumentFromReader fail, Error: %v", err.Error()))
			dataError = true
		} else {
			data = replaceWhite.ReplaceAllString(dataDoc.Text(), "")
		}
	})
	return data, dataError
}

func getDataFloat(titleData *goquery.Document, dataType string) (float64, bool) {
	var data float64
	dataError := false
	titleData.Find(cryptoFindFloat[dataType]).Each(func(_ int, selection *goquery.Selection) {
		dataString, err := goquery.OuterHtml(selection)
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("get data float outerhtml fail, Error: %v", err.Error()))
			dataError = true
		} else if dataDoc, err := goquery.NewDocumentFromReader(strings.NewReader(dataString)); err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("get data float NewDocumentFromReader fail, Error: %v", err.Error()))
			dataError = true
		} else if dataDoc.Text() != "" && dataDoc.Text() != "---" {
			data, err = strconv.ParseFloat(cryptoReg[dataType].ReplaceAllString(dataDoc.Text(), ""), 64)
			if err != nil {
				logger.Log(logger.LogLevel.Error, fmt.Sprintf("parsefloat fail, Error: %v", err.Error()))
				dataError = true
			}
		} else {
			dataError = true
		}
	})
	return data, dataError
}
