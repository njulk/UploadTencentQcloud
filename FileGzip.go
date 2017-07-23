package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

//判断是否存在此文件
func isExist(configurename, file string) (existed bool, errRet error) {
	_, errRet = os.Stat(file)
	if errRet != nil {
		if os.IsNotExist(errRet) {
			existed = false
			errRet = fmt.Errorf("文件%s不存在:%s", file, errRet.Error())
			return
		} else {
			existed = true
			errRet = fmt.Errorf("文件%s存在但是不能读写:%s", file, errRet.Error())
			log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
			return
		}
	} else {
		existed = true
		return
	}
}

func Compress(configurename string, srcFile string) (isselfGz, gzfileExisted bool, destFile string, errRet error) {
	if len(srcFile) >= 3 {
		tail := srcFile[(len(srcFile) - 3):len(srcFile)]
		if strings.EqualFold(tail, ".gz") {
			return true, false, srcFile, nil
		}
	}

	destFile = srcFile + ".gz"
	//查看压缩文件是否已经存在
	existed, errRet := isExist(configurename, destFile)
	if existed == true && errRet == nil {
		return false, true, destFile, nil
	}
	if existed == true && errRet != nil {
		log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
		return false, true, destFile, errRet
	}

	gzFile, err := os.Create(destFile)
	if err != nil {
		log.Error("配置文件%s:压缩时创建目标文件%s失败:%s\r\n", configurename, destFile, err.Error())
		err = fmt.Errorf("压缩时创建目标文件%s失败:%s", destFile, err.Error())
		return false, false, destFile, err
	}
	defer gzFile.Close()
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()
	rdFile, err := os.Open(srcFile)
	if err != nil {
		log.Error("配置文件%s:压缩时打开源文件%s失败:%s\r\n", configurename, srcFile, err.Error())
		err = fmt.Errorf("压缩时打开源文件%s失败:%s\r\n", srcFile, err.Error())
		return false, false, destFile, err
	}
	defer rdFile.Close()
	fileStat, err := os.Stat(srcFile)
	if err != nil {
		log.Error("配置文件%s:查看源文件%s的stat失败:%s\r\n", configurename, srcFile, err.Error())
		err = fmt.Errorf("查看源文件%s的stat失败:%s\r\n", srcFile, err.Error())
		return false, false, destFile, err
	}
	var mslice_size int64 = cosSliceSize
	var readData []byte = make([]byte, mslice_size, 2*mslice_size)
	var mslice_count int = int(math.Ceil((float64)(fileStat.Size()) / (float64)(mslice_size)))
	var offset int64 = 0
	for i := 0; i < mslice_count; i++ {
		readline, err := rdFile.ReadAt(readData, offset)
		if err != io.EOF && err != nil {
			log.Error("配置文件%s:压缩时读取源文件%s过程中出错:%s\r\n", configurename, srcFile, err.Error())
			err = fmt.Errorf("压缩时读取源文件%s过程中出错:%s\r\n", srcFile, err.Error())
			return false, false, destFile, err
		}
		_, err = gzWriter.Write(readData[0:readline])
		if err != nil {
			log.Error("配置文件%s:压缩时写入压缩文件%s出错:%s\r\n", configurename, destFile, err.Error())
			err = fmt.Errorf("压缩时写入压缩文件%s出错:%s", destFile, err.Error())
			return false, false, destFile, err
		}
		err = gzWriter.Flush()
		if err != nil {
			err = fmt.Errorf("压缩时写入压缩文件%s后flush出错:%s", destFile, err.Error())
			log.Error("配置文件%s:%s\r\n", configurename, err.Error())
			return false, false, destFile, err
		}
		offset += (int64)(readline)
	}
	return false, false, destFile, nil
}
