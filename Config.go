package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/larspensjo/config"
)

func getPara(configName string) (paras map[string]string, errRet error) {
	var configFile = flag.String("configfile", configName, "General configuration file")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	cfg, err := config.ReadDefault(*configFile)
	if err != nil {
		errRet = fmt.Errorf("%s", err.Error())
		fmt.Println(errRet.Error())
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
				}
			}
		}
	}
	if cfg.HasSection("DIR") {
		section, err := cfg.SectionOptions("DIR")
		if err == nil {
			for _, v := range section {
				options, err := cfg.String("DIR", v)
				if err == nil {
					paras[v] = options
				}
			}
		}
	}
	var parasRequire []string = []string{"appid", "bucket", "region", "secretId", "secretKey", "localPath"}
	for _, fi := range parasRequire {
		_, ok := paras[fi]
		if !ok {
			errRet = fmt.Errorf("para %s not find in %s", fi, configName)
			fmt.Println(errRet.Error())
			return
		}
	}
	return
}
