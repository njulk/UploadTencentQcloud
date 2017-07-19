package main

import (
	"compress/gzip"
	"io"
	"math"
	"os"
)

func Compress(srcFile string) (destFile string, errRet error) {
	destFile = srcFile + ".gz"
	gzFile, err := os.Create(destFile)
	if err != nil {
		log.Error("压缩时创建目标文件%s失败\r\n", destFile)
		return destFile, err
	}
	defer gzFile.Close()
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()
	rdFile, err := os.Open(srcFile)
	if err != nil {
		log.Error("压缩时打开源文件%s失败\r\n", srcFile)
		return destFile, err
	}
	defer rdFile.Close()
	fileStat, err := os.Stat(srcFile)
	if err != nil {
		log.Error("查看源文件%s的stat失败%s\r\n", srcFile)
		return destFile, err
	}
	var mslice_size int64 = 3145728
	var readData []byte = make([]byte, mslice_size, 2*mslice_size)
	var mslice_count int = int(math.Ceil((float64)(fileStat.Size()) / (float64)(mslice_size)))
	var offset int64 = 0
	for i := 0; i < mslice_count; i++ {
		readline, err := rdFile.ReadAt(readData, offset)
		if err != io.EOF && err != nil {
			log.Error("压缩时读取源文件%s过程中出错\r\n", srcFile)
			return destFile, err
		}
		gzWriter.Write(readData[0:readline])
		gzWriter.Flush()
		offset += mslice_size
	}
	return
}
