package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"xiaomipush"
)


var debugLog  *log.Logger
var package_name string
var message_title string
var message_body string
var pass_through int32
var LTMessageID string
var extra =  make(map[string]string)


const logFile string = "/data/log/goolink/xiaomi.log"   //日志文件

func initLog()*log.Logger{

	if !fileExists(logFile){
		logFile,err  := os.Create(logFile)
		defer logFile.Close()
		if err != nil {
			log.Fatalln("open file error !")
		}
		debugLog = log.New(logFile,"[Debug]",log.LstdFlags)
	}else{
		logFile, err := os.OpenFile(logFile, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		debugLog = log.New(logFile,"[Debug]",log.LstdFlags)
	}
	debugLog.SetPrefix("【fcm-push】")
	debugLog.SetFlags(log.LstdFlags | log.Lshortfile |log.LUTC)
	return debugLog
}

func writeLog(logInfo string){
	debugLog := initLog()
	debugLog.Println(logInfo,"\n\r")
}

func FatalfLog(format string, v ...interface{}) {
	debugLog := initLog()
	debugLog.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

//检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}


func strToMap(data string)map[string]string{
	var dataStr =[]byte(data)
	m := make(map[string]string)

	err := json.Unmarshal(dataStr, &m)
	if err != nil {
		FatalfLog("Umarshal failed: %v\n", err)
		return m
	}
	return m
}


func main() {
	t1 := time.Now()
	writeLog("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~xiaomi-push-begin~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	var contTimeStr string

	appSecret,regID,dataStr :=  os.Args[1],os.Args[2],os.Args[3]

	//写日志
	writeLog("appSecret="+appSecret)
	writeLog("regID="+regID)
	writeLog("dataStr="+dataStr)
	//转map
	dataMap :=  strToMap(dataStr)

	if _, ok := dataMap["push_mode"]; ok{
		if dataMap["push_mode"] == "1"{
			pass_through = 0
		}else{
			pass_through = 1
		}
	}

	//检查包名
	if _, ok := dataMap["package_name"]; ok{
		package_name = dataMap["package_name"]
	}else{
		fmt.Println("400|{\"reason\":\"no package_name\"}|0.0")
		writeLog("400|{\"reason\":\"no package_name\"}|0.0")
		return
	}

	//检查title
	if _, ok := dataMap["message_title"]; ok{
		message_title = dataMap["message_title"]
	}else{
		message_title = "This is xiaomi push title"
	}

	//检查body
	if _, ok := dataMap["message_body"]; ok{
		message_body = dataMap["message_body"]
	}else{
		message_body = "This is xiaomi push body"
	}

	//检查时间戳
	if _, ok := dataMap["LTMessageID"]; ok{
		LTMessageID = dataMap["LTMessageID"]
	}else{
		timeStamp := time.Now().Unix()
		LTMessageID = strconv.FormatInt(timeStamp, 10) //int64 to string
	}

	extra["LTMessageID"] = LTMessageID

	var client = xiaomipush.NewClient(appSecret, []string{package_name})

	var msg1 *xiaomipush.Message = xiaomipush.NewAndroidMessage(message_title, message_body).SetPayload("this is payload1")
	msg1.PassThrough = pass_through
	msg1.Extra = extra

	res,err := client.Send(context.Background(), msg1, regID)
	if err != nil {
		fmt.Println("405|{reson:"+err.Error()+"}|0.0")
		writeLog("405|{reson:"+err.Error()+"}|0.0")
		return
	}

	res2,err2 := client.Send2(context.Background(), msg1, regID)
	if err2 != nil {
		fmt.Println("Overseas push：405|{reson:"+err.Error()+"}|0.0")
		writeLog("Overseas push：405|{reson:"+err.Error()+"}|0.0")
		return
	}

	countTime := time.Since(t1)
	contTimeStr = shortDur(countTime)

	if res.Description == "成功"{
		fmt.Println("200|{\"reason\":\"sucess:"+res.Description+"\"}|Overseas push："+res2.Description+"|"+contTimeStr)
		writeLog("200|{\"reason\":\"sucess\"}|Overseas push："+res2.Description+"|"+contTimeStr)
	}else{
		fmt.Println("500|{reson:",res,"}|Overseas push：",res2,"|",contTimeStr)
		writeLog(res.Info)
		writeLog(res.Reason)
		writeLog(res.Description)
		writeLog("Overseas push："+res2.Info)
		writeLog("Overseas push："+res2.Reason)
		writeLog("Overseas push："+res2.Description)
	}

	writeLog("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~xiaomi-push-end~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

}


