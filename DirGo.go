package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//列出当前文件夹下的所有目录和文件
func ListDir(configurename string, dirpath string, selectSubdir bool) (files, dirs []string, errRet error) {
	dir, errRet := ioutil.ReadDir(dirpath)
	if errRet != nil {
		errRet = fmt.Errorf("读取目录时%s出现问题:%s", dirpath, errRet.Error())
		log.Error("配置文件%s:读取目录%s出现错误:%s\r\n", configurename, dirpath, errRet.Error())
		return
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
			subfiles, subdirs, err := ListDir(configurename, fi, true)
			if err != nil {
				log.Error("配置文件%s:%s\r\n", configurename, err.Error())
				errRet = err
				return
			}
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

//根据本地gz文件路径得出COS文件路径
func (cos *COS) extractDir(configurename string, localFile string) (filename string, errRet error) {
	if localFile == "" {
		errRet = fmt.Errorf("%s", "提取文件名出错，文件路径名空")
		log.Error("%s:提取文件名出错，文件路径名空\r\n", configurename)
		return filename, errRet
	}
	PathSep := string(os.PathSeparator)
	arrayFile := strings.Split(localFile, PathSep)
	filename = cos.uploadDir
	if len(arrayFile) == 1 {
		filename += "/" + arrayFile[0]
		return
	}
	if len(arrayFile) >= 2 {
		for i := 1; i < len(arrayFile); i++ {
			filename += "/" + arrayFile[i]
		}
	}
	return
}

//上传所有文件（根据一个所有文件的string数组）
func (cos *COS) uploadAllfiles(configurename string, allFiles []string, recordfile string) (errRet error) {

	recordData, errRet = getRecordData(configurename, recordfile)
	if errRet != nil {
		errRet = fmt.Errorf("获取记录失败%s", errRet.Error())
		log.Error("%s:%s\r\n", configurename, errRet.Error())
		return
	}
	if len(allFiles) > 5 {
		cos.startwork(configurename, 5, allFiles)
	} else {
		cos.startwork(configurename, len(allFiles), allFiles)
	}
	cosRecordName, errRet := cos.extractDir(configurename, recordfile)
	if errRet != nil {
		log.Error("%s:分析记录文件在cos的路径时出错\r\n", configurename)
		return
	}
	errRet = cos.uploadFile(configurename, recordfile, cosRecordName)
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s失败", recordfile)
	}

	return
}

//总的上传所有文件（根据本地目录）
func (cos *COS) uploadFromlocal(configurename string, filedir string, selectSubdir bool, recordfile string) (errRet error) {

	files, err := matchPath(configurename, filedir, selectSubdir)
	if err != nil {
		errRet = fmt.Errorf("%s", err.Error())
		log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return
	}
	if len(files) == 0 {
		log.Warn("配置文件:%s匹配文件个数为0,或许路径下忘记添加*\r\n", configurename)
		outlog.Warn("配置文件%s:匹配到文件个数为0,或许路径下忘记添加*\r\n", configurename)
		return
	}
	errRet = cos.uploadAllfiles(configurename, files, recordfile)
	if errRet != nil {
		log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return
	}
	return
}
