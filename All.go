package main

import (
	l4g "github.com/alecthomas/log4go"
)

//主函数的入口，总的综合函数，根据配置文件设置变量并且上传到相应的cos空间
func start() (errRet error) {
	var configureName string = "config.ini"
	paras, errRet := getPara(configureName)
	if errRet != nil {

		log.Error("获取参数失败:%s\r\n", errRet.Error())
		return
	}
	var objectcos *COS = new(COS)
	objectcos.appid = paras["appid"]
	objectcos.region = paras["region"]
	objectcos.bucket = paras["bucket"]
	objectcos.secretId = paras["secretId"]
	objectcos.secretKey = paras["secretKey"]
	recordTxtName = paras["recordTxtName"]
	initLog(l4g.DEBUG, paras["logPath"])
	defer log.Close()
	if paras["isRecurSub"] == "true" {
		errRet = objectcos.uploadFromlocal(paras["localPath"], true, recordTxtName)
		if errRet != nil {
			log.Error("执行上传文件失败:%s\r\n", errRet.Error())
		}
	} else {
		errRet = objectcos.uploadFromlocal(paras["localPath"], false, recordTxtName)
		if errRet != nil {
			log.Error("执行上传文件失败:%s\r\n", errRet.Error())
		}
	}
	return
}
