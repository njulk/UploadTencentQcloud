package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bitly/go-simplejson"
)

//创建目录
func (cos *COS) createDir(filepath string) (result string, errRet error) {
	url := cos.generateurl(filepath)
	sign, _ := cos.createSignature(filepath, false)
	var dir = make(map[string]string)
	dir["op"] = "create"
	dir["biz_attr"] = ""
	postdata, errRet := json.Marshal(dir)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	buffer := bytes.NewBuffer(postdata)
	var contenttype string = "application/json"
	tmpresult, errRet := cos.httppost(url, sign, contenttype, buffer)
	result = string(tmpresult)
	if errRet != nil {
		fmt.Println(errRet.Error())
		//return
	}
	return
}

//列出当前bucket中已经存在的目录
func (cos *COS) queryDir(rootDir string) (result []byte, subdirs []string, errRet error) {
	url := cos.generateurl(rootDir)
	var para string
	para = "?op=list&num=100"
	url += para
	sign, errRet := cos.createSignature(rootDir, false)
	if errRet != nil {
		return
	}
	result, errRet = cos.httpget(url, sign)
	if errRet != nil {
		return
	}
	jsonbody, errRet := simplejson.NewJson(result)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	jsondata, errRet := jsonbody.Get("data").Get("infos").Array()
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	for _, di := range jsondata {
		newbody, _ := di.(map[string]interface{})
		filesize := newbody["filesize"]
		filedir := newbody["name"]
		if filesize == nil {
			t := filedir.(string)
			subdirs = append(subdirs, rootDir+t)
			continue
		}
	}
	for _, v := range subdirs {
		_, subsubdirs, _ := cos.queryDir(v)
		for _, subv := range subsubdirs {
			subdirs = append(subdirs, subv)
		}
	}
	return
}

//判断当前目录是否存在，存在返回true，不存在返回false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

//列出当前文件夹下的所有目录和文件
func ListDir(dirpath string, selectSubdir bool) (files, dirs []string, errRet error) {
	dir, errRet := ioutil.ReadDir(dirpath)
	if errRet != nil {
		_, errRet := ioutil.ReadFile(dirpath)
		if errRet != nil {
			fmt.Println(errRet.Error())
			return nil, nil, errRet
		}
		files = []string{dirpath}
		return files, nil, nil
	}
	PathSep := string(os.PathSeparator)
	//PathSep := "/"
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

func (cos *COS) uploadAllfiles(allFiles []string, alreadyDirs map[string]string) (errRet error) {
	for _, file := range allFiles {
		fileDir, filepath := extractDir(file)
		//fmt.Println(fileDir)
		if _, ok := alreadyDirs[fileDir]; !ok {
			//fmt.Println(fileDir)
			_, errRet = cos.createDir(fileDir)
			if errRet != nil {
				return
			}
		} else {
			alreadyDirs[fileDir] = "ok"
		}
		_, errRet = Compress(file)
		if errRet != nil {
			return errRet
		}
		file += ".gz"
		filepath += ".gz"
		cos.uploadFile(file, filepath)
		errRet = os.Remove(file)
		if errRet != nil {
			fmt.Println(errRet.Error())
		}
	}
	return
}

func (cos *COS) uploadFromlocal(filedir string, selectSubdir bool) {
	if !Exist(filedir) {
		fmt.Println("%s：文件不存在", filedir)
		return
	}
	files, _, _ := ListDir(filedir, true)
	//fmt.Println(files)
	_, subdirs, _ := cos.queryDir("/")
	alreadyDirs := make(map[string]string)
	//fmt.Println(subdirs)
	for i := 0; i < len(subdirs); i++ {
		alreadyDirs[subdirs[i]] = "ok"
	}
	cos.uploadAllfiles(files, alreadyDirs)
	_, subdirs, _ = cos.queryDir("/")
	//fmt.Println(subdirs)
}
