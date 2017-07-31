package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

//http的post发送
func (cos *COS) httppost(configurename, url, sign, contenttype string, buffer *bytes.Buffer) (result []byte, errRet error) {
	var timeout = time.Duration(60 * 60 * time.Second)
	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		errRet = fmt.Errorf("http请求失败，message:%s", err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	req.Header.Set("Content-Type", contenttype)
	//req.Header.Set("Host", "gz.file.myqcloud.com")
	req.Header.Set("Authorization", sign)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		errRet = fmt.Errorf("http请求失败，message：%s", err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errRet = fmt.Errorf("http读取失败，message:%s", err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	return body, errRet
}

//http的get请求
func (cos *COS) httpget(configurename, url, sign string) (result []byte, errRet error) {
	var timeout = time.Duration(60 * 60 * time.Second)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errRet = fmt.Errorf("http请求失败，message:%s", err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	req.Header.Set("Authorization", sign)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		errRet = fmt.Errorf("http请求失败，message：%s", err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errRet = fmt.Errorf("http读取失败，message:%s", err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	return body, errRet
}
