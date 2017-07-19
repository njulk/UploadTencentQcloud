package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//列出当前文件夹下的所有目录和文件
func ListDir(dirpath string, selectSubdir bool) (files, dirs []string, errRet error) {
	dir, errRet := ioutil.ReadDir(dirpath)
	if errRet != nil {
		_, errRet := ioutil.ReadFile(dirpath)
		if errRet != nil {
			//fmt.Println(errRet.Error())
			log.Error("路径:%s既不是目录也不是文件\r\n", errRet.Error())
			return nil, nil, errRet
		}
		files = []string{dirpath}
		return files, nil, nil
	}
	PathSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() {
			dirs = append(dirs, dirpath+fi.Name()+PathSep)
		} else {
			files = append(files, dirpath+fi.Name())
		}
	}
	if selectSubdir {
		for _, fi := range dirs {
			subfiles, subdirs, _ := ListDir(fi, true)
			for _, tmpfiles := range subfiles {
				files = append(files, tmpfiles)
			}
			for _, tmpdirs := range subdirs {
				dirs = append(dirs, tmpdirs)
			}
		}
	}
	return
}

//根据本地文件路径得出COS文件路径以及COS文件上层目录路径
func extractDir(file string) (filedir string, filepath string) {
	PathSep := string(os.PathSeparator)
	i := strings.Split(file, PathSep)
	if len(i) <= 1 {
		return
	}
	if len(i) == 2 {
		return "/", "/" + i[1]
	}
	for x := 1; x < (len(i) - 1); x++ {
		filedir += "/" + i[x]
		filepath += "/" + i[x]
	}
	filedir += "/"
	filepath += "/" + i[len(i)-1]
	return
}

//上传所有文件（根据一个所有文件的string数组）
func (cos *COS) uploadAllfiles(allFiles []string, recordfile string) (errRet error) {

	recordData, errRet = getRecordData(recordfile)
	if errRet != nil {
		errRet = fmt.Errorf("获取记录失败%s", errRet.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	if len(allFiles) > 5 {
		cos.startwork(5, allFiles)
	} else {
		cos.startwork(len(allFiles), allFiles)
	}
	cosRecord := "/" + recordfile
	cos.uploadFile(recordfile, cosRecord)
	return
}

//总的上传所有文件（根据本地目录）
func (cos *COS) uploadFromlocal(filedir string, selectSubdir bool, recordfile string) (errRet error) {

	files, err := matchPath(filedir, selectSubdir)
	//fmt.Println(files)
	if err != nil {
		errRet = fmt.Errorf("匹配失败%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	if len(files) == 0 {
		return
	}
	errRet = cos.uploadAllfiles(files, recordfile)
	if errRet != nil {
		log.Error("%s\r\n", errRet.Error())
		return
	}
	return
}
