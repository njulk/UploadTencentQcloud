package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"

	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	/*COS提交数据的大小*/
	cosUploadURL string = "http://region.file.myqcloud.com/files/v2/"
	//cosUploadURL string = "http://www.google.com/"
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

func (cos *COS) generateurl(filepath string) string {
	tmp := strings.Replace(cosUploadURL, "region", cos.region, -1)
	return fmt.Sprintf("%s%s/%s%s", tmp, cos.appid, cos.bucket, filepath)
}

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
	//fmt.Println(result)
	mac := hmac.New(sha1.New, []byte(cos.secretKey))
	_, err = mac.Write([]byte(result))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	tmp := mac.Sum(nil)
	sign = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", string(tmp), result)))
	return sign, err
}

func getSecretKeyByAppId(appid string) (secretId, secretKey string, errRet error) {

	nowTs := time.Now().Unix()

	mac := hmac.New(sha1.New, []byte(cosCgiKey))
	mac.Write([]byte(fmt.Sprintf("%s&%d", cosCgiFrom, nowTs)))
	shaValue := string(mac.Sum(nil))
	sign := base64.StdEncoding.EncodeToString([]byte(shaValue))

	requestArray := url.Values{
		"Action":    {"HelperAppid2Key"},
		"appid":     {appid},
		"from":      {cosCgiFrom},
		"timestamp": {fmt.Sprintf("%d", nowTs)},
		"sign":      {sign}}
	requestUrl := fmt.Sprintf("%s?%s", cosSecretKeyUrl, requestArray.Encode())

	httpResponseData, err := httpget(requestUrl, "")
	js, err := simplejson.NewJson(httpResponseData)
	if err != nil {
		errRet = fmt.Errorf("结果解析失败,结果为:%s,错误为:%s", string(httpResponseData), err.Error())
		return
	}
	code, err := js.Get("code").Int()
	if err != nil {
		errRet = fmt.Errorf("结果解析失败(code),结果为:%s", string(httpResponseData))
		return
	}
	_, err = js.Get("message").String()
	if err != nil {
		errRet = fmt.Errorf("结果解析失败(message),结果为:%s", string(httpResponseData))
		return
	}
	if 0 != code {
		errRet = fmt.Errorf("取得上传Key失败,结果为:%s", string(httpResponseData))
		return
	}
	keys, err := js.Get("keys").Array()
	if err != nil {
		errRet = fmt.Errorf("取得上传Key失败(keys),结果为:%s", string(httpResponseData))
		return
	}

	for index, _ := range keys {
		status, err := js.Get("keys").GetIndex(index).Get("status").Int()
		if err != nil {
			errRet = fmt.Errorf("取得上传Key失败(status),结果为:%s", string(httpResponseData))
			return
		}
		if 2 == status {
			secretId, err = js.Get("keys").GetIndex(index).Get("secretId").String()
			if err != nil {
				errRet = fmt.Errorf("取得上传Key失败(secretId),结果为:%s", string(httpResponseData))
				return
			}
			secretKey, err = js.Get("keys").GetIndex(index).Get("secretKey").String()
			if err != nil {
				errRet = fmt.Errorf("取得上传Key失败(secretKey),结果为:%s", string(httpResponseData))
				return
			}

			if len(secretId) > 0 && len(secretKey) > 0 {
				break
			}
		}

	}
	if len(secretId) < 1 || len(secretKey) < 1 {
		errRet = fmt.Errorf("取得上传Key失败(secretId+secretKey),结果为:%s", string(httpResponseData))
		return
	}
	return secretId, secretKey, errRet
}
