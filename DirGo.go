package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

//列出当前文件夹下的所有目录和文件
/*func (cos *COS) ListDir(configurename string, dirpath string, selectSubdir bool) (files, dirs []string, errRet error) {
	dir, errRet := ioutil.ReadDir(dirpath)
	if errRet != nil {
		errRet = fmt.Errorf("读取目录时%s出现问题:%s", dirpath, errRet.Error())
		cos.log.Error("配置文件%s:读取目录%s出现错误:%s\r\n", configurename, dirpath, errRet.Error())
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
			subfiles, subdirs, err := cos.ListDir(configurename, fi, true)
			if err != nil {
				cos.log.Error("配置文件%s:%s\r\n", configurename, err.Error())
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
}*/

func (cos *COS) ListDir(configurename string, n *sync.WaitGroup, dirpath string, selectSubdir bool, dirschan chan string, errRets chan error) {
	defer n.Done()
	var dirs, files []string
	dir, errRet := ioutil.ReadDir(dirpath)
	if errRet != nil {
		errRet = fmt.Errorf("读取目录时%s出现问题:%s", dirpath, errRet.Error())
		errRets <- errRet
		cos.log.Error("配置文件%s:读取目录%s出现错误:%s\r\n", configurename, dirpath, errRet.Error())
		return
	}
	for _, fi := range dir {
		if fi.IsDir() {
			dirs = append(dirs, filepath.Join(dirpath, fi.Name()))
		} else {
			files = append(files, filepath.Join(dirpath, fi.Name()))
			dirschan <- filepath.Join(dirpath, fi.Name())
		}
	}
	if selectSubdir {
		for _, fi := range dirs {
			n.Add(1)
			go cos.ListDir(configurename, n, fi, true, dirschan, errRets)
		}
	}
	return
}

func (cos *COS) getDir(configurename, dirpath string, selectSubdir bool) (files []string, errRet error) {
	dirs := make(chan string, 30)
	errRets := make(chan error, 30)
	var wait sync.WaitGroup
	wait.Add(1)
	go cos.ListDir(configurename, &wait, dirpath, selectSubdir, dirs, errRets)
	go func() {
		wait.Wait()
		close(dirs)
		close(errRets)
	}()
	var filebool bool = true
	var errRetbool bool = true
	for {
		select {
		case filedir, ok := <-dirs:
			if !ok {
				filebool = false
				break
			}
			files = append(files, filedir)
			break
		case _, ok := <-errRets:
			if !ok {
				errRetbool = false
				break
			}
			errRet = fmt.Errorf("提取文件目录出现错误，具体查看Log日志")
			cos.log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
			break
		}

		if !filebool && !errRetbool {
			return files, errRet
		}

	}

}

//根据本地gz文件路径得出COS文件路径
func (cos *COS) extractDir(configurename string, localFile string) (filename string, errRet error) {
	if localFile == "" {
		errRet = fmt.Errorf("%s", "提取文件名出错，文件路径名空")
		cos.log.Error("%s:提取文件名出错，文件路径名空\r\n", configurename)
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
func (cos *COS) uploadAllfiles(configurename string, allFiles []string) (errRet error) {

	cos.recordData, errRet = cos.getRecordData(configurename)
	if errRet != nil {
		errRet = fmt.Errorf("获取记录失败%s", errRet.Error())
		cos.log.Error("%s:%s\r\n", configurename, errRet.Error())
		return
	}
	num := runtime.NumCPU()
	num = num*2 + 1
	if len(allFiles) > num {
		errRet = cos.startwork(configurename, num, allFiles)
	} else {
		errRet = cos.startwork(configurename, len(allFiles), allFiles)
	}
	if errRet != nil {
		cos.log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		cos.outlog.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return
	}
	cosRecordName, errRet := cos.extractDir(configurename, cos.recordTxtName)
	if errRet != nil {
		cos.log.Error("%s:分析记录文件在cos的路径时出错\r\n", configurename)
		return
	}
	errRet = cos.uploadFile(configurename, cos.recordTxtName, cosRecordName)
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s失败", cos.recordTxtName)
	}

	return
}

//总的上传所有文件（根据本地目录）
func (cos *COS) uploadFromlocal(configurename string, filedir string, selectSubdir bool) (errRet error) {

	files, err := cos.matchPath(configurename, filedir, selectSubdir)
	if err != nil {
		errRet = fmt.Errorf("%s", err.Error())
		cos.log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return
	}
	if len(files) == 0 {
		cos.log.Warn("配置文件:%s匹配文件个数为0,或许路径下忘记添加*\r\n", configurename)
		cos.outlog.Warn("配置文件%s:匹配到文件个数为0,或许路径下忘记添加*\r\n", configurename)
		return
	}
	errRet = cos.uploadAllfiles(configurename, files)
	if errRet != nil {
		cos.log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return
	}
	return
}
