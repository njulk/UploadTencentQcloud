package main

import (
	l4g "github.com/alecthomas/log4go"
)

//主函数的入口，总的综合函数，根据配置文件设置变量并且上传到相应的cos空间
func start(configureName string) (errRet error) {
	var objectcos *COS = new(COS)
	objectcos.initCOS()
	objectcos.outlog.AddFilter("logger", l4g.FINE, l4g.NewConsoleLogWriter())
	defer objectcos.outlog.Close()
	paras, errRet := objectcos.getPara(configureName)
	if errRet != nil {
		objectcos.outlog.Error("配置文件%s:获取参数失败:%s\r\n", configureName, errRet.Error())
		return
	}
	objectcos.log.AddFilter("file", l4g.FINE, l4g.NewFileLogWriter(paras["logPath"], false))
	defer objectcos.log.Close()
	errRet = objectcos.setPara(configureName, paras)
	if errRet != nil {
		objectcos.log.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		objectcos.outlog.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		return
	}
	if paras["isRecurSub"] == "true" {
		errRet = objectcos.uploadFromlocal(configureName, paras["localPath"], true)
		if errRet != nil {
			objectcos.log.Error("配置文件%s:程序结束:%s\r\n", configureName, errRet.Error())
			objectcos.outlog.Error("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志\r\n", configureName, errRet.Error())
			return
		}
	} else {
		errRet = objectcos.uploadFromlocal(configureName, paras["localPath"], false)
		if errRet != nil {
			objectcos.log.Error("配置文件%s:程序结束:%s\r\n", configureName, errRet.Error())
			objectcos.outlog.Error("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志\r\n", configureName, errRet.Error())
			return
		}
	}
	objectcos.outlog.Info("配置文件%s:程序结束,无误,上传完毕\r\n", configureName)
	return
}

func (objectcos *COS) setPara(configureName string, paras map[string]string) (errRet error) {
	objectcos.appid = paras["appid"]
	id, key, errRet := objectcos.getSecretKeyByAppId(configureName, objectcos.appid)
	if errRet != nil {
		objectcos.log.Error("%s\r\n", errRet.Error())
		return
	}
	objectcos.region = paras["region"]
	objectcos.bucket = paras["bucket"]
	objectcos.secretId = id
	objectcos.secretKey = key
	objectcos.uploadDir = paras["uploadDir"]
	objectcos.recordTxtName = paras["recordTxtName"]
	localIp, errRet := objectcos.getLocalIp(configureName)
	if errRet != nil {
		objectcos.outlog.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		objectcos.log.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		return
	}
	errRet = objectcos.detectCosDir(configureName, localIp)
	if errRet != nil {
		objectcos.outlog.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		objectcos.log.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		return
	}
	return
}
