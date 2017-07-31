package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	l4g "github.com/alecthomas/log4go"
)

const (
	/*COS提交数据的大小*/
	cosUploadURL string = "http://region.file.myqcloud.com/files/v2/"
	/*简单上传文件*/
	cosPostMaxSize int64 = 20971520
	//cosPostMaxSize int64 = 209
	/*分片大小，单位为 Byte*/
	cosSliceSize int64 = 3145728
	/*签名过期的时间*/
	signExpiration int64  = 3600
	userAgent      string = "cos-golang-sdk-1.0"
)

type CreateDir struct {
	Op       string `json:"op"`
	Biz_attr string `json:"biz_attr"`
}

/*
COS上传需要的SecretKey
*/
type SecretKey struct {
	appid     string
	secretId  string
	secretKey string
}

type localInfo struct {
	log           l4g.Logger
	outlog        l4g.Logger
	mutex         sync.RWMutex
	recordData    map[string]string
	recordTxtName string
}

/*COS对象*/
type COS struct {
	SecretKey
	localInfo
	region    string
	bucket    string
	uploadDir string
}

/*COS上面的文件信息*/
type COSFile struct {
	code       int
	message    string
	filelen    int64
	filesize   int64
	sha        string
	source_url string
}

var cosCgiKey string = "f9a3b98f20e1a0a7ca97ce7319b717cb"
var cosCgiFrom string = "cdn"
var cosSecretKeyUrl string = "http://cosapi4.qcloud.com/api.php"

var logMutex sync.Mutex
var configMutex sync.Mutex
var configSize int

//初始化结构体
func (cos *COS) initCOS() {
	cos.log = make(l4g.Logger)
	cos.outlog = make(l4g.Logger)
	cos.recordData = make(map[string]string)
}

//为屏幕打印日志添加滤波器
func addFilter(configname string, cos *COS) {
	logMutex.Lock()
	cos.outlog.AddFilter(configname, l4g.FINE, l4g.NewConsoleLogWriter())
	logMutex.Unlock()
}

//为文本打印日志添加滤波器
func addLogFilter(configname, logpath string, cos *COS) {
	logMutex.Lock()
	cos.log.AddFilter(configname, l4g.FINE, l4g.NewFileLogWriter(logpath, false))
	logMutex.Unlock()
}

//产生基本的url
func (cos *COS) generateurl(filepath string) string {
	tmp := strings.Replace(cosUploadURL, "region", cos.region, -1)
	return fmt.Sprintf("%s%s/%s%s", tmp, cos.appid, cos.bucket, filepath)
}

//创造签名
func (cos *COS) createSignature(configurename string, filePath string, singleUse bool) (sign string, err error) {
	var cur = time.Now().Unix()
	var expiration = cur + 3600
	if singleUse {
		expiration = 0
	}
	var result string
	var fileId string
	fileId = fmt.Sprintf("/%s/%s%s", cos.appid, cos.bucket, filePath)
	result = fmt.Sprintf("a=%s&b=%s&k=%s&e=%d&t=%d&r=%d&f=%s", cos.appid, cos.bucket, cos.secretId, expiration, cur, rand.Int(), fileId)
	mac := hmac.New(sha1.New, []byte(cos.secretKey))
	_, err = mac.Write([]byte(result))
	if err != nil {
		cos.log.Error("配置文件%s:cos文件%s签名写入出错:%s", configurename, filePath, err.Error())
		err = fmt.Errorf("cos文件%s签名写入出错:%s", filePath, err.Error())
		return
	}
	tmp := mac.Sum(nil)
	sign = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", string(tmp), result)))
	return sign, err
}

//获取ip地址
func (cos *COS) getLocalIp(configurename string) (Ip string, errRet error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		errRet = fmt.Errorf("获取本机IP错误:%s", err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	} else {
		if len(addrs) == 0 {
			errRet = fmt.Errorf("获取到0个IP")
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return
		} else {
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String(), errRet
					}
				}
			}
			errRet = fmt.Errorf("没有合适的IP")
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return
		}
	}
}
