package main

import (
	"Common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)
const logFile string = "/yourPath/huaweiPush.log"   //日志文件
const token_session_path string= "/yourPath/session/"  //存放access_token的文件
const huawei_oauth2_url string = "https://login.cloud.huawei.com/oauth2/v2/token"  

func init(){
	Common.InitLog(logFile)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeLog(info string){
	Common.WriteLog(info,logFile)
}


type auth struct{
	grant_type         string
	client_secret      string
	client_id          string
}

func(auth2 *auth) readToken(filename string)string{
	b, err := ioutil.ReadFile(token_session_path+"huawei_session_"+filename)
	if err != nil {
		fmt.Print(err)
	}
	return string(b)
}

func (auth2 *auth)writeToken(filename,data string){
	d1 := []byte(data)
	err := ioutil.WriteFile(token_session_path+"huawei_session_"+filename, d1, 0644)
	check(err)
}

func (auth2 *auth)getNewToken()(string,string){
	form := auth2.defaultForm()
	res,err := auth2.httpPost(huawei_oauth2_url,form)
	if res == ""{
		return "",err
	}else{
		data := Common.JsonToMap(res)
		return data["access_token"].(string),err
	}

}

func (auth2 *auth)getHuaweiOuth2Token()(string,string){
	timestamp := time.Now().Unix()
	if auth2.checkSessionExit(){
		jsonStr := auth2.readToken(auth2.client_id)
		data := Common.JsonToMap(jsonStr)
		old_timestamp := int64((data["timestamp"]).(float64))

		if (old_timestamp + 3000) < timestamp {
			//fmt.Println("token已过期")
			writeLog("access_token已过期,重新生成")
			//token已过期
			token,err := auth2.getNewToken()
			writeLog("重新生成的access_token="+token)
			if token != ""{
				data := make(map[string]interface{})
				data["token"] = token
				data["timestamp"] = timestamp
				jsonStr := Common.MapTojson(data)
				auth2.writeToken(auth2.client_id,jsonStr)

				return token,""
			}else{
				return "",err
			}

		}else{

			//token没有过期
			writeLog("access_token没有过期,复用")
			return data["token"].(string),""
		}

	}else{
		writeLog("没有access_token，第一次生成")
		token,err := auth2.getNewToken()
		writeLog("access_token="+token)
		if token != ""{
			data := make(map[string]interface{})
			data["token"] = token
			data["timestamp"] = timestamp
			jsonStr := Common.MapTojson(data)
			auth2.writeToken(auth2.client_id,jsonStr)

			return token,""
		}else{
			return "",err
		}
	}
}

func (auth2 *auth)checkSessionExit()bool{
	path := token_session_path+"huawei_session_"+auth2.client_id
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func (auth2 *auth) defaultForm() url.Values{
	form := url.Values{}
	form.Add("grant_type", auth2.grant_type)
	form.Add("client_secret", auth2.client_secret)
	form.Add("client_id", auth2.client_id)
	return form

}

func (auth2 *auth)httpPost(url string ,form url.Values) (string,string){
	resp, err := http.Post(url,"application/x-www-form-urlencoded",strings.NewReader(form.Encode()))
	if resp.StatusCode == 200{
		check(err)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		check(err)
		return string(body),""
	}else{
		return "",resp.Status
	}

}


type HuweiPush struct{
	access_token string
	nsp_svc string
	nsp_ts string
	device_token_list string
	payload string
}

func (push *HuweiPush)defaultForm() url.Values{
	form := url.Values{}
	form.Add("access_token", push.access_token)
	form.Add("nsp_svc", push.nsp_svc)
	form.Add("nsp_ts", push.nsp_ts)
	form.Add("device_token_list", push.device_token_list)
	form.Add("payload", push.payload)
	//fmt.Println(form)
	return form
}

func (push *HuweiPush)httpPost(url string ,form url.Values) (string,string){

	resp, err := http.Post(url,"application/x-www-form-urlencoded",strings.NewReader(form.Encode()))
	if resp.StatusCode == 200{
		check(err)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		check(err)
		return string(body),""
	}else{
		return "",resp.Status
	}

}

func (push *HuweiPush)getPushUrl(appId string)string{
	nsp_ctx_map := make(map[string]interface{})
	nsp_ctx_map["ver"] = "1"
	nsp_ctx_map["appId"] = appId
	nsp_ctx_urlencodeStr := url.QueryEscape(Common.MapTojson(nsp_ctx_map))
	huawei_push_url := "https://api.push.hicloud.com/pushsend.do?nsp_ctx="+nsp_ctx_urlencodeStr
	return huawei_push_url
}

func getHuaweiPayload(dataMap map[string]interface{})string{
	push_mode,content,title,intent,appPkgName := dataMap["push_mode"].(string),dataMap["message_body"].(string),dataMap["message_title"].(string),dataMap["popupActivity"].(string),dataMap["log_sys_id"].(string)

	payload := make(map[string]interface{})
	hps := make(map[string]interface{})
	msg := make(map[string]interface{})
	body :=make(map[string]string)
	if push_mode == "0"{
		push_mode = "1"  //华为推送 1 透传异步消息
		writeLog("走透传异步消息")
		body["content"] = content
		body["title"] = title
		msg["type"] = push_mode
		msg["body"] = body
		hps["msg"] = msg
		payload["hps"] = hps

		payloadJsoonStr := Common.MapTojson(payload)
		return payloadJsoonStr
	}else{
		push_mode = "3"	//华为推送 3 系统通知栏异步消息
		writeLog("走系统通知栏异步消息")
		body["content"] = content
		body["title"] = title
		msg["type"] = push_mode
		msg["body"] = body
		action := make(map[string]interface{})
		param := make(map[string]interface{})
		param["intent"] = intent
		param["appPkgName"] = appPkgName

		action["type"] = 1
		action["param"] = param
		msg["action"] = action
		hps["msg"] = msg
		hps["ext"] = dataMap
		payload["hps"] = hps

		payloadJsoonStr := Common.MapTojson(payload)
		return payloadJsoonStr

	}
}



func main(){
	t1 := time.Now()
	writeLog("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~huaweiPush-push-begin~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	var contTimeStr string

	client_secret,registrationToken,dataStr :=  os.Args[1],os.Args[2],os.Args[3]

	//写日志
	writeLog("client_secret="+client_secret)
	writeLog("registrationToken="+registrationToken)
	writeLog("dataStr="+dataStr)

	dataMap :=  Common.JsonToMap(dataStr)
	client_id := dataMap["appkey"].(string)
	auth2 := auth{"client_credentials",client_secret,client_id}

	access_token,_ := auth2.getHuaweiOuth2Token()
	stamp := time.Now().Unix()
	timestamp := strconv.FormatInt(stamp, 10)
	var tokenList  [1]string

	tokenList[0] = registrationToken

	tokenListJson, _ := json.Marshal(tokenList)

	payloadJsoonStr := getHuaweiPayload(dataMap)

	push := HuweiPush{access_token,"openpush.message.api.send",timestamp,string(tokenListJson),payloadJsoonStr}

	huawei_push_url := push.getPushUrl(client_id)
	resp, _ := push.httpPost(huawei_push_url,push.defaultForm())
	countTime := time.Since(t1)
	contTimeStr = Common.ShortDur(countTime)
	respMap := Common.JsonToMap(resp)
	if _, ok := respMap["msg"]; ok{
		if respMap["msg"] == "Success" || respMap["code"]=="80000000"{
			fmt.Println("200|{\"reason\":\"sucess:"+resp+"\"}|"+contTimeStr)
			writeLog("200|{\"reason\":\"sucess:"+resp+"\"}|"+contTimeStr)
		}else{
			fmt.Println("400|reason",resp,contTimeStr)
			writeLog("400|{\"reason\":\""+resp+"\"}|"+contTimeStr)
		}
	}else{
		fmt.Println("500|reason",resp,contTimeStr)
		writeLog("500|{\"reason\":\""+resp+"\"}|"+contTimeStr)
	}

	writeLog("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~huaweiPush-push-end~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
}
