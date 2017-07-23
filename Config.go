package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/larspensjo/config"
)

//更换用户写的读写文件路径的/或\使得在适应Linux下和windows下的写法
func dirSuitSys(dir string) (dst string) {
	dst = filepath.FromSlash(dir)
	dst = strings.Replace(dst, "\\", string(os.PathSeparator), -1)
	return
}

//更换用户的上传路径的\必须为/不能是\\
func updateUploadDir(dir string) (dst string) {
	return strings.Replace(dir, "\\", "/", -1)
}

//根据配置文件获取各个参数
func getPara(configName string) (paras map[string]string, errRet error) {
	paras = make(map[string]string)
	var parasRequire []string = []string{"appid",
		"bucket",
		"uploadDir",
		"region",
		"localPath",
		"isRecurSub",
		"recordTxtName",
		"logPath"}
	cfg, err := config.ReadDefault(configName)
	if err != nil {
		errRet = fmt.Errorf("读取配置文件错误:%s,配置文件为:%s", err.Error(), configName)
		outlog.Error("%s\r\n", errRet.Error())
		return
	}
	for _, sectionName := range []string{"COS", "DIR"} {
		if cfg.HasSection(sectionName) {
			if section, err := cfg.SectionOptions(sectionName); err == nil {
				for _, v := range section {
					if options, err := cfg.String(sectionName, v); err == nil {
						paras[v] = options
						//fmt.Println(options)
					} else {
						errRet = fmt.Errorf("获取%s目录下参数:%s,错误:%s,配置文件为:%s", sectionName, v, err.Error(), configName)
						outlog.Error("%s\r\n", errRet.Error())
						return
					}
				}
			} else {
				errRet = fmt.Errorf("获取配置中的%s目录错误,错误:%s,配置文件为:%s", sectionName, err.Error(), configName)
				outlog.Error("%s\r\n", errRet.Error())
				return
			}
		} else {
			errRet = fmt.Errorf("配置中未配置%s项目,配置文件为:%s", sectionName, configName)
			outlog.Error("%s\r\n", errRet.Error())
			return
		}
	}
	for _, fi := range parasRequire {
		if _, ok := paras[fi]; !ok {
			errRet = fmt.Errorf("配置文件中不存在配置:%s,配置文件为:%s", fi, configName)
			outlog.Error("%s\r\n", errRet.Error())
			return
		} else {
			if fi == "localPath" || fi == "recordTxtName" || fi == "logPath" {
				paras[fi] = dirSuitSys(paras[fi])
			}
			if fi == "uploadDir" {
				paras[fi] = updateUploadDir(paras[fi])
			}
		}
	}

	if paras["appid"] == "" || paras["bucket"] == "" || paras["region"] == "" {
		errRet = fmt.Errorf("appid,bucket,region参数有缺失")
		outlog.Error("%s\r\n", errRet.Error())
		return
	}
	if paras["isRecurSub"] != "false" {
		paras["isRecurSub"] = "true"
	}
	errRet = detectUpdate(configName, paras["recordTxtName"], true)
	if errRet != nil {
		return
	}
	errRet = detectUpdate(configName, paras["logPath"], false)
	if errRet != nil {
		return
	}
	return
}

//用于判断日志和记录文件是否合理
func detectUpdate(configname string, srcfile string, isRecord bool) (errRet error) {
	if srcfile == "" {
		errRet = fmt.Errorf("配置里记录文件或日志文件不能为空")
		outlog.Error("配置文件%s:%s\r\n", configname, errRet.Error())
		return
	}
	fi, err := os.Stat(srcfile)
	if err != nil {
		if os.IsNotExist(err) {
			fc, errCreate := os.Create(srcfile)
			if errCreate != nil {
				errRet = fmt.Errorf("创建文件失败,文件:%s,错误:%s", srcfile, errCreate.Error())
				outlog.Error("配置文件%s:%s\r\n", configname, errRet.Error())
				return
			} else {
				defer fc.Close()
				if isRecord {
					_, errRet = fc.WriteString("localfile" + "                       " + "cosfile" + "\r\n")
					if errRet != nil {
						log.Error("%s:记录文件%s记录数据时出错:%s\r\n", configname, srcfile, errRet.Error())
						errRet = fmt.Errorf("记录文件%s记录数据时出错:%s", srcfile, errRet.Error())
						return errRet
					}
				} else {
					return
				}
			}
		} else {
			errRet = fmt.Errorf("读取文件信息失败,文件:%s,错误:%s", srcfile, err.Error())
			outlog.Error("配置文件%s:%s\r\n", configname, errRet.Error())
			return
		}
	} else {
		if fi.IsDir() {
			errRet = fmt.Errorf("文件:%s配置错误,不能是目录", srcfile)
			outlog.Error("配置文件%s:%s\r\n", configname, errRet.Error())
			return
		} else {
			return
		}
	}
	return
}

//检测配置中cos的uploadDir是否设置正确
func (cos *COS) detectCosDir(configurename string, localIp string) (errRet error) {
	if cos.uploadDir[(len(cos.uploadDir)-1):len(cos.uploadDir)] != "/" {
		errRet = fmt.Errorf("配置文件里的uploadDir格式不对，最右边缺少/")
		log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return
	}
	if cos.uploadDir[0:1] != "/" {
		errRet = fmt.Errorf("配置文件里的uploadDir格式不对，最左边缺少/")
		log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return
	}
	cos.uploadDir = cos.uploadDir + localIp
	return
}
