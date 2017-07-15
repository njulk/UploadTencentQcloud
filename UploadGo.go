package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"math"
	"mime/multipart"
	"os"

	"github.com/bitly/go-simplejson"
)

//"io/ioutil"

func (cos *COS) smallFileupload(localFilePath string, filePath string) (cosFile string, errRet error) {
	//cosFile.code = -1
	url := cos.generateurl(filePath)
	sign, err := cos.createSignature(filePath, false)
	if err != nil {
		fmt.Errorf("生成签名失败", err.Error())
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentWriter, err := bodyWriter.CreateFormFile("filecontent", localFilePath)
	if err != nil {
		errRet = fmt.Errorf("内容初始化失败(filecontent)", err.Error())
	}
	localFileReader, err := os.Open(localFilePath)
	if err != nil {
		errRet = fmt.Errorf("本地文件打开失败", err.Error())
	}
	defer localFileReader.Close()
	_, err = io.Copy(contentWriter, localFileReader)
	if err != nil {
		errRet = fmt.Errorf("内容拷贝初始化失败", err.Error())
	}
	if _, err = localFileReader.Seek(0, os.SEEK_SET); err != nil {
		errRet = fmt.Errorf(err.Error())
	}
	shaObject := sha1.New()
	if _, err = io.Copy(shaObject, localFileReader); err != nil {
		errRet = fmt.Errorf("本地文件sha1失败", err.Error())
	}
	filesha := fmt.Sprintf("%x", shaObject.Sum(nil))
	contentWriter, _ = bodyWriter.CreateFormField("op")
	contentWriter.Write([]byte("upload"))

	contentWriter, _ = bodyWriter.CreateFormField("sha")
	contentWriter.Write([]byte(filesha))

	contentWriter, _ = bodyWriter.CreateFormField("insertOnly")
	contentWriter.Write([]byte("0"))

	contentType := bodyWriter.FormDataContentType()
	resp, err := cos.httppost(url, sign, contentType, &buffer)
	cosFile = string(resp)
	return
}

func (cos *COS) uploadinit(localFilePath string, filePath string) (session string, filesize, slice_size int64, errRet error) {
	url := cos.generateurl(filePath)
	sign, err := cos.createSignature(filePath, false)
	//data, _ := ioutil.ReadFile(localFilePath)
	//filesize = int64(len(data))
	fileinfo, err := os.Stat(localFilePath)
	if err != nil {
		errRet = fmt.Errorf("文件信息获取失败,%s", err.Error())
		return
	}
	filesize = fileinfo.Size()
	if err != nil {
		fmt.Errorf("生成签名失败", err.Error())
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, err := bodyWriter.CreateFormField("op")
	contentwriter.Write([]byte("upload_slice_init"))

	contentwriter, err = bodyWriter.CreateFormField("filesize")
	contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))

	contentwriter, err = bodyWriter.CreateFormField("slice_size")
	contentwriter.Write([]byte(fmt.Sprintf("%d", cosSliceSize)))

	contentwriter, err = bodyWriter.CreateFormField("insertOnly")
	contentwriter.Write([]byte(fmt.Sprintf("%d", 0)))

	contentType := bodyWriter.FormDataContentType()
	resp, _ := cos.httppost(url, sign, contentType, &buffer)

	js, _ := simplejson.NewJson(resp)

	code, err := js.Get("code").Int()
	if err != nil {
		errRet = fmt.Errorf("解析失败()code，结果为%s", err.Error())
		return
	}
	if code != 0 {
		errRet = fmt.Errorf("cos上传失败,结果为%s", string(resp))
	}

	session, err = js.Get("data").Get("session").String()
	if err != nil {
		errRet = fmt.Errorf("cos初始化失败，结果为:%s", string(resp))
	}

	slice_size, err = js.Get("data").Get("slice_size").Int64()
	return
}

func (cos *COS) upload_slice_finish(filepath, session string, filesize int64) (cosfile string, errRet error) {
	url := cos.generateurl(filepath)
	sign, _ := cos.createSignature(filepath, false)
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, _ := bodyWriter.CreateFormField("op")
	contentwriter.Write([]byte("upload_slice_finish"))

	contentwriter, _ = bodyWriter.CreateFormField("session")
	contentwriter.Write([]byte(session))

	contentwriter, _ = bodyWriter.CreateFormField("filesize")
	contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))

	contentType := bodyWriter.FormDataContentType()
	resp, _ := cos.httppost(url, sign, contentType, &buffer)
	cosfile = string(resp)
	return
}

func (cos *COS) upload_slice_data(localFilePath string, filesize int64, filepath, session string, slice_size int64) (isOk bool, errRet error) {
	url := cos.generateurl(filepath)
	sign, _ := cos.createSignature(filepath, false)
	localFileReader, err := os.Open(localFilePath)
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()

	var offset, sliceNowNumber int64
	var readData []byte = make([]byte, slice_size, 2*slice_size)
	sliceCount := int64(math.Ceil(float64(filesize) / float64(slice_size)))
	for sliceNowNumber < sliceCount {
		buffer.Truncate(0)
		offset = sliceNowNumber * slice_size
		readlen, _ := localFileReader.ReadAt(readData, offset)
		if readlen == 0 && err == io.EOF {
			break
		}
		sliceNowNumber++
		contentwriter, _ := bodyWriter.CreateFormField("op")
		contentwriter.Write([]byte("upload_slice_data"))

		contentwriter, _ = bodyWriter.CreateFormField("filecontent")
		contentwriter.Write(readData[0:readlen])

		contentwriter, _ = bodyWriter.CreateFormField("session")
		contentwriter.Write([]byte(session))

		contentwriter, _ = bodyWriter.CreateFormField("offset")
		contentwriter.Write([]byte(fmt.Sprintf("%d", offset)))

		contentType := bodyWriter.FormDataContentType()
		cos.httppost(url, sign, contentType, &buffer)

	}
	return true, errRet
}

func (cos *COS) largeFileupload(localFilePath string, filepath string) (result string) {
	session, filesize, slicesize, _ := cos.uploadinit(localFilePath, filepath)
	cos.upload_slice_data(localFilePath, filesize, filepath, session, slicesize)
	result, _ = cos.upload_slice_finish(filepath, session, filesize)
	//fmt.Println(result)
	return
}

func (cos *COS) uploadFile(localFilePath string, filePath string) {
	fileinfo, _ := os.Stat(localFilePath)
	filesize := fileinfo.Size()
	if filesize <= cosPostMaxSize {
		cos.smallFileupload(localFilePath, filePath)
	} else {
		cos.largeFileupload(localFilePath, filePath)
	}
}
