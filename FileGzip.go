package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
)

func Compress(srcFile string) (destFile string, errRet error) {
	destFile = srcFile + ".gz"
	gzFile, err := os.Create(destFile)
	defer gzFile.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()
	rdFile, err := os.Open(srcFile)
	defer rdFile.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	content, _ := ioutil.ReadAll(rdFile)
	_, err = gzWriter.Write(content)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = gzWriter.Flush()
	if err != nil {
		fmt.Println(err)
		return
	}
	//gzWriter.Close()
	return

}
