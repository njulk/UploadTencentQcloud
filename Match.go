package main

import (
	"os"
	"path/filepath"
)

//根据路径匹配到相应要上传的文件
func matchPath(pattern string, selectsub bool) (matches []string, errRet error) {
	files, errRet := filepath.Glob(pattern)
	if errRet != nil {
		log.Error("Glob匹配出问题:%s", errRet.Error())
		return nil, errRet
	}
	if len(files) == 0 {
		log.Warn("匹配到0个文件\r\n")
		return
	}
	var allfiles []string
	var isFirst bool = true
	for i := 0; i < len(files); i++ {
		fi, _ := os.Stat(files[i])
		if fi.IsDir() {
			if selectsub || isFirst {
				PathSep := string(os.PathSeparator)
				slen := len([]rune(files[i]))
				if files[i][slen-1:slen] != PathSep {
					files[i] = files[i] + PathSep
				}
			} else {
				continue
			}
		}
		if selectsub || isFirst {
			subfiles, _, errRet := ListDir(files[i], selectsub)
			if errRet != nil {
				log.Error("迭代子目%s录列表失败\r\n", files[i])
				continue
			}
			for _, v := range subfiles {
				allfiles = append(allfiles, v)
			}
		} else {
			break
		}
		isFirst = false
	}
	return allfiles, errRet
}
