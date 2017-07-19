package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/larspensjo/config"
)

//根据配置文件获取各个参数
func getPara(configName string) (paras map[string]string, errRet error) {
	var configFile = flag.String("configfile", configName, "General configuration file")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	cfg, err := config.ReadDefault(*configFile)
	if err != nil {
		errRet = fmt.Errorf("读取配置文件错误:%s\r\n", err.Error())
		log.Error(errRet.Error())
		return nil, errRet
	}
	paras = make(map[string]string)
	if cfg.HasSection("COS") {
		section, err := cfg.SectionOptions("COS")
		if err == nil {
			for _, v := range section {
				options, err := cfg.String("COS", v)
				if err == nil {
					paras[v] = options
				} else {
					log.Warn("获取COS目录下参数错误:%s\r\n", paras[v])
				}
			}
		} else {
			log.Warn("获取参数COS目录错误\r\n")
		}
	}
	if cfg.HasSection("DIR") {
		section, err := cfg.SectionOptions("DIR")
		if err == nil {
			for _, v := range section {
				options, err := cfg.String("DIR", v)
				if err == nil {
					paras[v] = options
				} else {
					log.Warn("获取DIR目录下参数错误:%s\r\n", paras[v])
				}
			}
		} else {
			log.Warn("获取参数DIR目录错误\r\n")
		}
	}
	var parasRequire []string = []string{"appid", "bucket", "region", "localPath", "isRecurSub", "recordTxtName", "logPath"}
	for _, fi := range parasRequire {
		_, ok := paras[fi]
		if !ok {
			errRet = fmt.Errorf("para %s not find in %s", fi, configName)
			log.Error(errRet.Error())
			return
		}
	}
	if paras["appid"] == "" || paras["bucket"] == "" || paras["region"] == "" {
		log.Error("appid,bucket,region参数有缺失,不能为空\r\n")
		errRet = fmt.Errorf("appid,bucket,region参数有缺失")
		return
	}
	if paras["isRecurSub"] != "false" {
		paras["isRecurSub"] = "true"
	}
	if paras["recordTxtName"] == "" {
		paras["recordTxtName"] = "record.txt"
	}
	if paras["logPath"] == "" {
		paras["logPath"] = "log.txt"
	}
	filep, errf := os.Open(paras["localPath"])
	_, errd := ioutil.ReadDir(paras["localPath"])
	if err != nil && errd != nil {
		log.Error("配置文件中本地路径localPath有错\r\n")
		errRet = fmt.Errorf("配置文件中本地路径localPath有错")
		return
	}
	if errf == nil {
		filep.Close()
	}
	return
}
