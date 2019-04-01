package Common

import (
	"bytes"
	"crypto/rc4"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)


type Rc4 struct{
	Rc4Key string
}

//加密
func (r *Rc4)Rc4Encode(json_str string)string{
	rc4Key := r.Rc4Key
	var key []byte = []byte(rc4Key) //初始化用于加密的KEY
	rc4obj1, _ := rc4.NewCipher(key) //返回 Cipher
	rc4str1 := []byte(json_str)  //需要加密的字符串
	plaintext := make([]byte, len(rc4str1))
	rc4obj1.XORKeyStream(plaintext, rc4str1)
	//base64加密
	input := []byte(plaintext)
	encodeString := base64.StdEncoding.EncodeToString(input)
	return encodeString
}
//解密
func (r *Rc4)Rc4Decode(encodeStr string)string{
	rc4Key := r.Rc4Key
	//base64解密
	decodeBytes, err := base64.StdEncoding.DecodeString(encodeStr)
	if err != nil {
		log.Fatalln(err)
	}
	//rc4解密
	c, err := rc4.NewCipher([]byte(rc4Key))
	if err != nil {
		log.Fatalln(err)
	}
	src := []byte(decodeBytes)
	dst := make([]byte, len(src))
	c.XORKeyStream(dst, src)
	return string(dst)

}

func HttpPostJson(url string,dataStr string) string{
	var jsonStr = []byte(dataStr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Timeout", "0.1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func GetPulicIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}



func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func JsonToMap(jsonStr string)map[string]interface{}{
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &m)
	CheckErr(err)
	return m
}

func MapTojson(datas map[string]interface{})string{
		jsonStr, _ := json.Marshal(datas)
		return string(jsonStr)
}

var debugLog  *log.Logger

func InitLog(logFile string)*log.Logger{

	if !FileExists(logFile){
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
	debugLog.SetPrefix("【push】")
	debugLog.SetFlags(log.LstdFlags | log.Lshortfile |log.LUTC)
	return debugLog
}

//检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func WriteLog(logInfo,filename string){
	debugLog := InitLog(filename)
	debugLog.Println(logInfo,"\n\r")
}

func FatalfLog(format,filename string, v ...interface{}) {
	debugLog := InitLog(filename)
	debugLog.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func ShortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}




