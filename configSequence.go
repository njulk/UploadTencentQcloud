package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//获取目录的绝对路径
func getAbsDir(filePath string) (absPath string, errRet error) {
	absPath, errRet = filepath.Abs(filePath)
	if errRet != nil {
		errRet = fmt.Errorf("获取路径%s的绝对路径失败:%s", filePath, errRet.Error())
		return
	}
	return
}

//获取两者的最小值
func min(len1, len2 int) int {
	if len1 < len2 {
		return len1
	} else {
		return len2
	}
}

//判断两个目录之间是否包含
func dirContain(dir1, dir2 string) (isContain bool, errRet error) {
	if dir1 == "" || dir2 == "" {
		errRet = fmt.Errorf("文件目录为空")
		isContain = false
		return
	}
	subdir1 := strings.Split(dir1, string(os.PathSeparator))
	subdir2 := strings.Split(dir2, string(os.PathSeparator))
	length := min(len(subdir1), len(subdir2))
	for i := 0; i < length; i++ {
		if subdir1[i] != subdir2[i] {
			isContain = false
			return
		}
	}
	isContain = true
	return

}

//在一堆目录里选出可并行的目录和串行运行的目录
func getDiff(dirs []string) (parallel, serial []string, errRet error) {
	num := len(dirs)
	isContain := make([]bool, num)
	for x := 0; x < num; x++ {
		isContain[x] = false
	}
	for i := 0; i < num; i++ {
		if isContain[i] == true {
			continue
		}
		for j := i + 1; j < num; j++ {
			if isContain[j] == true {
				continue
			}
			tmpContain, err := dirContain(dirs[i], dirs[j])
			if err != nil {
				errRet = err
				return
			} else {
				if tmpContain == true {
					isContain[j] = true
					serial = append(serial, dirs[j])
				}
				if isContain[i] == false {
					parallel = append(parallel, dirs[i])
					isContain[i] = true
				}

			}

		}
		if isContain[i] == false {
			parallel = append(parallel, dirs[i])
		}
	}
	return
}

//如果是文件的话获取上层目录，如果是目录的话获取本身
func (cos *COS) getPreLevDir(dir string) (dst string, errRet error) {
	matches, errRet := filepath.Glob(dir)
	if errRet != nil {
		cos.outlog.Error("匹配目录%s出问题:%s", dst, errRet.Error())
		return
	}
	if len(matches) == 0 {
		dst = ""
		return
	}
	if dir[len(dir)-1:len(dir)] == "*" {
		dst, _ = filepath.Split(matches[0])
	} else {
		dst = matches[0]
		fileinfo, err := os.Stat(dst)
		if err != nil {
			errRet = fmt.Errorf("查询%s文件信息失败:%s", dst, err.Error())
			return
		}
		if !fileinfo.IsDir() {
			dst, _ = filepath.Split(matches[0])
		}
	}
	dst, errRet = getAbsDir(dst)
	return

}

//获取配置文件里的配置文件名与本地路径之间的对应关系
func mapConfigLocatepath(files []string) (mapRelation map[string]([]string), errRet error) {
	cos := new(COS)
	cos.initCOS()
	mapRelation = make(map[string][]string)
	addFilter("totalConfig", cos)
	defer cos.outlog.Close()
	for i := 0; i < len(files); i++ {
		para, err := cos.getPara(files[i])
		if err != nil {
			errRet = fmt.Errorf("配置文件%s:%s", files[i], err.Error())
			cos.outlog.Error("配置文件%s:%s", files[i], errRet.Error())
			continue
		}
		para["localPath"], err = cos.getPreLevDir(para["localPath"])
		if err != nil {
			errRet = err
			cos.outlog.Error("配置文件%s:%s", files[i], errRet.Error())
			continue
		}
		if para["localPath"] != "" {
			mapRelation[para["localPath"]] = append(mapRelation[para["localPath"]], files[i])
		} else {
			cos.outlog.Warn("配置文件%s的localpath匹配到零个文件", files[i])
		}
	}
	return
}

//将配置文件根据本地目录分为并行和串行的部分
func testConfig(configs []string) (parallelConfig, serialConfig []string, errRet error) {
	info, err := mapConfigLocatepath(configs)
	if err != nil {
		errRet = err
	}
	var files []string
	for path, _ := range info {
		files = append(files, path)
	}
	para, seri, err := getDiff(files)
	if err != nil {
		errRet = err
	}
	for _, v := range para {
		parallelConfig = append(parallelConfig, info[v][0])
		for x := 1; x < len(info[v]); x++ {
			serialConfig = append(serialConfig, info[v][x])
		}
	}
	for _, v := range seri {
		for x := 0; x < len(info[v]); x++ {
			serialConfig = append(serialConfig, info[v][x])
		}
	}
	return
}
