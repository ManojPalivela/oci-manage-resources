package commonutils

import (
	"log"
	"os"
	"strconv"
	"time"
)

//write log to specified file
func WriteLog(logFile string, message string) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Println(message)
}

//get unique string based on current time in GMT
func GetTimeStamp() string {
	repoTZ, _ := time.LoadLocation("GMT")
	return GetTimeStampWithTZ(repoTZ)
}

//get unique string based on current time in specified time zonee
func GetTimeStampWithTZ(tz *time.Location) string {
	now := time.Now()
	year := strconv.Itoa(now.In(tz).Year())
	var month string
	var day string
	var hour string
	var minute string
	var second string
	var milliSecond string
	//hour
	//minute
	//second
	//millis
	if int(now.In(tz).Month()) <= 9 {
		month = "0" + strconv.Itoa(int(now.In(tz).Month()))
	} else {
		month = strconv.Itoa(int(now.In(tz).Month()))
	}
	if now.In(tz).Day() <= 9 {
		day = "0" + strconv.Itoa(now.In(tz).Day())
	} else {
		day = strconv.Itoa(now.In(tz).Day())
	}

	if int(now.In(tz).Hour()) <= 9 {
		hour = "0" + strconv.Itoa(int(now.In(tz).Hour()))
	} else {
		hour = strconv.Itoa(int(now.In(tz).Hour()))
	}

	if int(now.In(tz).Minute()) <= 9 {
		minute = "0" + strconv.Itoa(int(now.In(tz).Minute()))
	} else {
		minute = strconv.Itoa(int(now.In(tz).Minute()))
	}

	if int(now.In(tz).Second()) <= 9 {
		minute = "0" + strconv.Itoa(int(now.In(tz).Second()))
	} else {
		minute = strconv.Itoa(int(now.In(tz).Second()))
	}

	millis := now.In(tz).Nanosecond()
	millis = millis / 1000000
	if millis <= 9 {
		milliSecond = "00" + strconv.Itoa(int(now.In(tz).Second()))
	} else if millis <= 99 {
		milliSecond = "0" + strconv.Itoa(int(now.In(tz).Second()))
	} else {
		milliSecond = strconv.Itoa(millis)
	}

	return year + month + day + "_" + hour + minute + second + "_" + milliSecond

}
