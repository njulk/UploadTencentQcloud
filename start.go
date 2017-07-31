package main

import (
	"fmt"
	"sync"
)

//具体执行一个配置文件的入口，总的综合函数，根据配置文件设置变量并且上传到相应的cos空间
func start(configureName string, n *sync.WaitGroup, result chan error) (errRet error) {
	if result != nil {
		defer n.Done()
	}
	var objectcos *COS = new(COS)
	objectcos.initCOS()
	addFilter(configureName, objectcos)
	defer objectcos.outlog.Close()
	paras, errRet := objectcos.getPara(configureName)
	if errRet != nil {
		objectcos.outlog.Error("配置文件%s:获取参数失败:%s", configureName, errRet.Error())
		errRet = fmt.Errorf("配置文件%s:获取参数失败:%s", configureName, errRet.Error())
		if result != nil {
			result <- errRet
		}
		return
	}
	addLogFilter(configureName, paras["logPath"], objectcos)
	defer objectcos.log.Close()
	errRet = objectcos.setPara(configureName, paras)
	if errRet != nil {
		objectcos.log.Error("配置文件%s:%s", configureName, errRet.Error())
		objectcos.outlog.Error("配置文件%s:%s", configureName, errRet.Error())
		errRet = fmt.Errorf("配置文件%s:%s", configureName, errRet.Error())
		if result != nil {
			result <- errRet
		}
		return
	}
	if paras["isRecurSub"] == "true" {
		errRet = objectcos.uploadFromlocal(configureName, paras["localPath"], true)
		if errRet != nil {
			objectcos.log.Error("配置文件%s:程序结束:%s", configureName, errRet.Error())
			objectcos.outlog.Error("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志", configureName, errRet.Error())
			errRet = fmt.Errorf("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志", configureName, errRet.Error())
			if result != nil {
				result <- errRet
			}
			return
		}
	} else {
		errRet = objectcos.uploadFromlocal(configureName, paras["localPath"], false)
		if errRet != nil {
			objectcos.log.Error("配置文件%s:程序结束:%s", configureName, errRet.Error())
			objectcos.outlog.Error("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志", configureName, errRet.Error())
			errRet = fmt.Errorf("配置文件%s:程序结束,出现错误:%s，具体请查看Log日志", configureName, errRet.Error())
			if result != nil {
				result <- errRet
			}
			return
		}
	}
	objectcos.outlog.Info("配置文件%s:程序结束,无误,上传完毕\r\n", configureName)
	return
}

//设置参数
func (objectcos *COS) setPara(configureName string, paras map[string]string) (errRet error) {
	objectcos.appid = paras["appid"]
	objectcos.region = paras["region"]
	objectcos.bucket = paras["bucket"]
	objectcos.secretId = paras["secretId"]
	objectcos.secretKey = paras["secretKey"]
	objectcos.uploadDir = paras["uploadDir"]
	objectcos.recordTxtName = paras["recordTxtName"]
	localIp, errRet := objectcos.getLocalIp(configureName)
	if errRet != nil {
		objectcos.outlog.Error("配置文件%s:%s", configureName, errRet.Error())
		objectcos.log.Error("配置文件%s:%s", configureName, errRet.Error())
		return
	}
	errRet = objectcos.detectCosDir(configureName, localIp)
	if errRet != nil {
		objectcos.outlog.Error("配置文件%s:%s", configureName, errRet.Error())
		objectcos.log.Error("配置文件%s:%s", configureName, errRet.Error())
		return
	}
	return
}
