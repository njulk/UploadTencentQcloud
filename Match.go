package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func matchPath(pattern string, selectsub bool) (matches []string, errRet error) {
	files, errRet := filepath.Glob(pattern)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return nil, errRet
	}
	var allfiles []string
	for i := 0; i < len(files); i++ {
		fi, _ := os.Stat(files[i])
		if fi.IsDir() {
			if selectsub {
				PathSep := string(os.PathSeparator)
				slen := len([]rune(files[i]))
				if files[i][slen-1:slen] != PathSep {
					files[i] = files[i] + PathSep
				}
			} else {
				continue
			}
		}
		subfiles, _, errRet := ListDir(files[i], selectsub)
		if errRet != nil {
			fmt.Println(errRet.Error())
			return nil, errRet
		}
		for _, v := range subfiles {
			allfiles = append(allfiles, v)
		}
	}
	return allfiles, errRet
}
