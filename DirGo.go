package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bitly/go-simplejson"
)

//创建目录
func (cos *COS) createDir(filepath string) string {
	url := cos.generateurl(filepath)
	sign, _ := cos.createSignature(filepath, false)
	var dir = make(map[string]string)
	dir["op"] = "create"
	dir["biz_attr"] = ""
	postdata, errRet := json.Marshal(dir)
	if errRet != nil {
		fmt.Println("request body error")
	}
	buffer := bytes.NewBuffer(postdata)
	var contenttype string = "application/json"
	result, errRet := cos.httppost(url, sign, contenttype, buffer)
	return string(result)
}

//列出当前bucket中已经存在的目录
func (cos *COS) queryDir(rootDir string) (result []byte, subdirs []string, errRet error) {
	url := cos.generateurl(rootDir)
	var para string
	para = "?op=list&num=100"
	url += para
	sign, _ := cos.createSignature(rootDir, false)
	result, err := cos.httpget(url, sign)
	if err != nil {
		errRet = fmt.Errorf("%s", err.Error())
	}
	//fmt.Println(string(result))
	jsonbody, _ := simplejson.NewJson(result)
	jsondata, _ := jsonbody.Get("data").Get("infos").Array()

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

//列出当前文件夹下的所有目录和文件
func ListDir(dirpath string, selectSubdir bool) (files, dirs []string, err error) {
	dir, err := ioutil.ReadDir(dirpath)
	if err != nil {
		_, errRet := ioutil.ReadFile(dirpath)
		if errRet != nil {
			return nil, nil, errRet
		}
		files = []string{dirpath}
		return files, nil, err
	}
	//PathSep := string(os.PathSeparator)
	PathSep := "/"
	for _, fi := range dir {
		if fi.IsDir() {
			dirs = append(dirs, dirpath+fi.Name()+PathSep)
		} else {
			files = append(files, dirpath+fi.Name())
		}
	}
	if selectSubdir {
		//fmt.Println(dirs)
		for _, fi := range dirs {
			//fmt.Println(fi)
			subfiles, subdirs, _ := ListDir(fi, true)
			//fmt.Println(subfiles)
			//fmt.Println(subdirs)
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

func extractDir(file string) (result string) {
	i := strings.Split(file, "/")
	if len(i) <= 2 {
		return "/"
	}
	for x := 1; x < (len(i) - 1); x++ {
		result += "/" + i[x]
	}
	return result
}

func (cos *COS) uploadAllfiles(allFiles []string, alreadyDirs map[string]string) {
	for _, file := range allFiles {
		fileDir := extractDir(file)
		if _, ok := alreadyDirs[fileDir]; !ok {
			cos.createDir(fileDir)
		}
		cos.uploadFile(file, file)
	}
}

func (cos *COS) uploadFromlocal(filedir string, selectSubdir bool) {
	files, _, _ := ListDir(filedir, selectSubdir)
	_, subdirs, _ := cos.queryDir("/")
	//fmt.Println(files)
	//fmt.Println(dirs)
	//fmt.Println(subdirs)
	alreadyDirs := make(map[string]string)
	for i := 0; i < len(subdirs); i++ {
		alreadyDirs[subdirs[i]] = "ok"
	}
	cos.uploadAllfiles(files, alreadyDirs)
}
