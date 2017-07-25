package main

import (
	"fmt"
	"os"
	"path/filepath"
)

//判断目录是否带有最后一个/，没有的话自动添上
func addDirTail(dir string) (result string) {
	PathSep := string(os.PathSeparator)
	slen := len(dir)
	if (rune)(dir[slen-1]) != (rune)(PathSep[0]) {
		result = dir + PathSep
	} else {
		result = dir
	}
	return result
}

//根据路径匹配到相应要上传的文件
func (cos *COS) matchPath(configurename string, pattern string, selectsub bool) (matches []string, errRet error) {
	files, errRet := filepath.Glob(pattern)
	if errRet != nil {
		cos.log.Error("配置文件%s:Glob匹配出问题:%s\r\n", configurename, errRet.Error())
		errRet = fmt.Errorf("Glob匹配样式%s出问题:%s", pattern, errRet.Error())
		return nil, errRet
	}
	if len(files) == 0 {
		return
	}
	var allfiles []string
	for i := 0; i < len(files); i++ {
		fi, err := os.Stat(files[i])
		if err != nil {
			cos.log.Error("配置文件%s:查询文件%s信息出错:%s\r\n", configurename, files[i], err.Error())
			errRet = fmt.Errorf("匹配文件%s时查询文件信息出错:%s", files[i], err.Error())
			return
		}
		if fi.IsDir() {
			files[i] = addDirTail(files[i])
		}

		if selectsub {
			if fi.IsDir() {
				subfiles, err := cos.getDir(configurename, files[i], selectsub)
				if err != nil {
					cos.log.Error("配置文件%s:%s\r\n", configurename, err.Error())
					errRet = fmt.Errorf("%s", err.Error())
					return
				} else {
					for _, v := range subfiles {
						allfiles = append(allfiles, v)
					}
				}

			} else {
				allfiles = append(allfiles, files[i])
			}

		} else {
			if fi.IsDir() {
				continue
			} else {
				allfiles = append(allfiles, files[i])
			}
		}
	}
	return allfiles, errRet
}
