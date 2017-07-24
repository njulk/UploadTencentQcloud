package main

import (
	l4g "github.com/alecthomas/log4go"
)

//主函数的入口，总的综合函数，根据配置文件设置变量并且上传到相应的cos空间
func start(configureName string) (errRet error) {
	outlog.AddFilter("logger", l4g.FINE, l4g.NewConsoleLogWriter())
	defer outlog.Close()
	paras, errRet := getPara(configureName)
	if errRet != nil {
		outlog.Error("配置文件%s:获取参数失败:%s\r\n", configureName, errRet.Error())
		return
	}
	initLog(l4g.DEBUG, paras["logPath"])
	defer log.Close()
	/*errRet = log.Warn("Log日志打开，测试是否可以读写\r\n")
	if errRet != nil {
		outlog.Error("log日志写入失败,请查看文件%s权限是否合适\r\n", paras["logPath"])
		return
	}*/
	var objectcos *COS = new(COS)
	errRet = objectcos.setPara(configureName, paras)
	if errRet != nil {
		log.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		outlog.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		return
	}
	if paras["isRecurSub"] == "true" {
		errRet = objectcos.uploadFromlocal(configureName, paras["localPath"], true, recordTxtName)
		if errRet != nil {
			log.Error("配置文件%s:程序结束:%s\r\n", configureName, errRet.Error())
			outlog.Error("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志\r\n", configureName, errRet.Error())
			return
		}
	} else {
		errRet = objectcos.uploadFromlocal(configureName, paras["localPath"], false, recordTxtName)
		if errRet != nil {
			log.Error("配置文件%s:程序结束:%s\r\n", configureName, errRet.Error())
			outlog.Error("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志\r\n", configureName, errRet.Error())
			return
		}
	}
	outlog.Info("配置文件%s:程序结束,无误,上传完毕\r\n", configureName)
	return
}

func (objectcos *COS) setPara(configureName string, paras map[string]string) (errRet error) {
	objectcos.appid = paras["appid"]
	id, key, errRet := getSecretKeyByAppId(configureName, objectcos.appid)
	if errRet != nil {
		log.Error("%s\r\n", errRet.Error())
		return
	}
	objectcos.region = paras["region"]
	objectcos.bucket = paras["bucket"]
	objectcos.secretId = id
	objectcos.secretKey = key
	objectcos.uploadDir = paras["uploadDir"]
	recordTxtName = paras["recordTxtName"]
	localIp, errRet := getLocalIp(configureName)
	if errRet != nil {
		outlog.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		log.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		return
	}
	errRet = objectcos.detectCosDir(configureName, localIp)
	if errRet != nil {
		outlog.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		log.Error("配置文件%s:%s\r\n", configureName, errRet.Error())
		return
	}
	return
}
