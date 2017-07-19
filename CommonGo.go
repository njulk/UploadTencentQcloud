package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
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

/*COS对象*/
type COS struct {
	SecretKey
	region string
	bucket string
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
var recordData map[string]string = make(map[string]string)
var log l4g.Logger = make(l4g.Logger)

var recordTxtName string
var mutex sync.Mutex

//产生基本的url
func (cos *COS) generateurl(filepath string) string {
	tmp := strings.Replace(cosUploadURL, "region", cos.region, -1)
	return fmt.Sprintf("%s%s/%s%s", tmp, cos.appid, cos.bucket, filepath)
}

//创造签名
func (cos *COS) createSignature(filePath string, singleUse bool) (sign string, err error) {
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
		log.Error("签名写入出错:%s\r\n", err.Error())
		return
	}
	tmp := mac.Sum(nil)
	sign = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", string(tmp), result)))
	return sign, err
}

//初始化log环境
func initLog(lv l4g.Level, logPath string) {
	log.AddFilter("file", lv, l4g.NewFileLogWriter(logPath, false))
}
