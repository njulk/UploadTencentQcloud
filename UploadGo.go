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

//上传小文件到对应的cos
func (cos *COS) smallFileupload(localFilePath string, filePath string) (Resbody string, errRet error) {
	url := cos.generateurl(filePath)
	sign, err := cos.createSignature(filePath, false)
	if err != nil {
		errRet = fmt.Errorf("生成签名失败", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentWriter, err := bodyWriter.CreateFormFile("filecontent", localFilePath)
	if err != nil {
		errRet = fmt.Errorf("内容初始化失败(filecontent)", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	localFileReader, err := os.Open(localFilePath)
	if err != nil {
		errRet = fmt.Errorf("本地文件打开失败", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	defer localFileReader.Close()
	_, err = io.Copy(contentWriter, localFileReader)
	if err != nil {
		errRet = fmt.Errorf("内容拷贝失败", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	if _, err = localFileReader.Seek(0, os.SEEK_SET); err != nil {
		errRet = fmt.Errorf("seek起点失败:%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	shaObject := sha1.New()
	if _, err = io.Copy(shaObject, localFileReader); err != nil {
		errRet = fmt.Errorf("本地文件sha1失败", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	filesha := fmt.Sprintf("%x", shaObject.Sum(nil))
	contentWriter, err = bodyWriter.CreateFormField("op")
	if err != nil {
		errRet = fmt.Errorf("创建属性op出现错误:%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	_, err = contentWriter.Write([]byte("upload"))
	if err != nil {
		errRet = fmt.Errorf("创建属性upload出现错误:%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}

	contentWriter, err = bodyWriter.CreateFormField("sha")
	if err != nil {
		errRet = fmt.Errorf("创建属性sha出现错误:%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	_, err = contentWriter.Write([]byte(filesha))
	if err != nil {
		errRet = fmt.Errorf("写入sha属性内容出现错误:%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	contentWriter, err = bodyWriter.CreateFormField("insertOnly")
	if err != nil {
		errRet = fmt.Errorf("创建insertOnly属性出现错误:%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	_, err = contentWriter.Write([]byte("0"))
	if err != nil {
		errRet = fmt.Errorf("写入insertOnly属性内容出现错误:%s", err.Error())
		log.Error("%s\r\n", errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := httppost(url, sign, contentType, &buffer)
	if errRet != nil {
		log.Error("http发送小文件%s时候出现错误\r\n", localFilePath)
		return
	}
	Resbody = string(resp)
	return
}

//上传大文件到对应的cos的初始化
func (cos *COS) uploadinit(localFilePath string, filePath string) (session string, filesize, slice_size int64, errRet error) {
	url := cos.generateurl(filePath)
	sign, errRet := cos.createSignature(filePath, false)
	if errRet != nil {
		log.Error("上传大文件%s初始化时候出错:%s\r\n", localFilePath, errRet.Error())
		return
	}
	fileinfo, errRet := os.Stat(localFilePath)
	if errRet != nil {
		log.Error("查询大文件%s信息失败:%s\r\n", localFilePath, errRet.Error())
		return
	}
	filesize = fileinfo.Size()
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, errRet := bodyWriter.CreateFormField("op")
	if errRet != nil {
		log.Error("httpbody创建属性op出错:%s\r\n", errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte("upload_slice_init"))
	if errRet != nil {
		log.Error("httpbody写入属性op出错:%s\r\n", errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("filesize")
	if errRet != nil {
		log.Error("httpbody创建属性filesize出错:%s\r\n", errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))
	if errRet != nil {
		log.Error("httpbody写入属性filesize出错:%s\r\n", errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("slice_size")
	if errRet != nil {
		log.Error("httpbody创建属slice_size出错:%s\r\n", errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", cosSliceSize)))
	if errRet != nil {
		log.Error("httpbody写入属性slice_size出错:%s\r\n", errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("insertOnly")
	if errRet != nil {
		log.Error("httpbody创建属性insertOnly出错:%s\r\n", errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", 0)))
	if errRet != nil {
		log.Error("httpbody写入属性insertOnly出错:%s\r\n", errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := httppost(url, sign, contentType, &buffer)
	if errRet != nil {
		log.Error("http上传大文件初始化发送失败:%s\r\n", errRet.Error())
		return
	}
	js, errRet := simplejson.NewJson(resp)
	if errRet != nil {
		log.Error("http大文件初始化响应数据json格式初始化失败:%s\r\n", errRet.Error())
		return
	}
	code, errRet := js.Get("code").Int()
	if errRet != nil {
		log.Error("响应数据获得code码失败:%s\r\n", errRet.Error())
		return
	}
	if code != 0 {
		log.Error("响应数据获得code码不为0,获得数据失败\r\n")
		return
	}
	session, errRet = js.Get("data").Get("session").String()
	if errRet != nil {
		log.Error("响应数据获得session码失败:%s\r\n", errRet.Error())
		return
	}
	slice_size, errRet = js.Get("data").Get("slice_size").Int64()
	if errRet != nil {
		log.Error("响应数据获得slice_size失败:%s\r\n", errRet.Error())
		return
	}
	return
}

//上传大文件的尾部到cos
func (cos *COS) upload_slice_finish(filepath, session string, filesize int64) (Resbody string, errRet error) {
	url := cos.generateurl(filepath)
	sign, errRet := cos.createSignature(filepath, false)
	if errRet != nil {
		log.Error("上传大文件尾部创建签名失败:%s\r\n", errRet.Error())
		return
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, errRet := bodyWriter.CreateFormField("op")
	if errRet != nil {
		log.Error("httpbody创建op属性失败:%s\r\n", errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte("upload_slice_finish"))
	if errRet != nil {
		log.Error("httpbody写入op属性内容upload_slice_finish失败:%s\r\n", errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("session")
	if errRet != nil {
		log.Error("httpbody创建session属性失败:%s\r\n", errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(session))
	if errRet != nil {
		log.Error("httpbody写入session属性内容失败:%s\r\n", errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("filesize")
	if errRet != nil {
		log.Error("httpbody创建session属性失败:%s\r\n", errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))
	if errRet != nil {
		log.Error("httpbody写入filesize属性内容失败:%s\r\n", errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := httppost(url, sign, contentType, &buffer)
	if errRet != nil {
		log.Error("http发送大文件尾部失败\r\n")
		return
	}
	Resbody = string(resp)
	return
}

//上传大文件的具体数据到cos
func (cos *COS) upload_slice_data(localFilePath string, filesize int64, filepath, session string, slice_size int64) (isOk bool, errRet error) {
	url := cos.generateurl(filepath)
	sign, errRet := cos.createSignature(filepath, false)
	if errRet != nil {
		log.Error("大文件内容创建签名失败:%s\r\n", errRet.Error())
		fmt.Println(errRet.Error())
		return
	}
	localFileReader, errRet := os.Open(localFilePath)
	if errRet != nil {
		log.Error("打开大文件%s失败:%s\r\n", localFilePath, errRet.Error())
		return
	}
	defer localFileReader.Close()
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()

	var offset, sliceNowNumber int64
	var readData []byte = make([]byte, slice_size, 2*slice_size)
	sliceCount := int64(math.Ceil(float64(filesize) / float64(slice_size)))
	for sliceNowNumber < sliceCount {
		buffer.Truncate(0)
		offset = sliceNowNumber * slice_size
		readlen, errRet := localFileReader.ReadAt(readData, offset)
		if errRet != nil && errRet != io.EOF {
			log.Error("读取大文件内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		if readlen == 0 && errRet == io.EOF {
			break
		}
		sliceNowNumber++
		contentwriter, errRet := bodyWriter.CreateFormField("op")
		if errRet != nil {
			log.Error("httpbody创建op属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte("upload_slice_data"))
		if errRet != nil {
			log.Error("httpbody写入upload_slice_data属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		contentwriter, errRet = bodyWriter.CreateFormField("filecontent")
		if errRet != nil {
			log.Error("httpbody创建filecontent属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write(readData[0:readlen])
		if errRet != nil {
			log.Error("httpbody写入filecontent属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}

		contentwriter, errRet = bodyWriter.CreateFormField("session")
		if errRet != nil {
			log.Error("httpbody创建session属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte(session))
		if errRet != nil {
			log.Error("httpbody写入session属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		contentwriter, errRet = bodyWriter.CreateFormField("offset")
		if errRet != nil {
			log.Error("httpbody创建offset属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", offset)))
		if errRet != nil {
			log.Error("httpbody写入offset属性内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
		contentType := bodyWriter.FormDataContentType()
		_, errRet = httppost(url, sign, contentType, &buffer)
		if errRet != nil {
			log.Error("http发送大文件内容失败:%s\r\n", errRet.Error())
			return false, errRet
		}
	}
	return true, errRet
}

//上传大文件的总的综合函数
func (cos *COS) largeFileupload(localFilePath string, filepath string) (result string, errRet error) {
	session, filesize, slicesize, err := cos.uploadinit(localFilePath, filepath)
	if err != nil {
		log.Error("上传大文件%s初始化失败:%s\r\n", localFilePath, err.Error())
		errRet = fmt.Errorf("%s", err.Error())
		return
	}
	isOk, errRet := cos.upload_slice_data(localFilePath, filesize, filepath, session, slicesize)
	if isOk != true {
		log.Error("上传大文件%s数据内容失败:%s\r\n", localFilePath, errRet.Error())
		return
	}
	result, errRet = cos.upload_slice_finish(filepath, session, filesize)
	if errRet != nil {
		log.Error("上传大文件%s数据内容失败:%s\r\n", localFilePath, errRet.Error())
		return
	}
	return
}

//上传大小文件的综合函数
func (cos *COS) uploadFile(localFilePath string, filePath string) error {
	fileinfo, errRet := os.Stat(localFilePath)
	if errRet != nil {
		log.Error("查询文件%s信息失败:%s\r\n", localFilePath, errRet.Error())
		return errRet
	}
	filesize := fileinfo.Size()
	if filesize <= cosPostMaxSize {
		_, errRet = cos.smallFileupload(localFilePath, filePath)
		if errRet != nil {
			log.Error("上传小文件%s出问题:%s\r\n", localFilePath, errRet.Error())
		}
		return errRet
	} else {
		_, errRet = cos.largeFileupload(localFilePath, filePath)
		if errRet != nil {
			log.Error("上传大文件%s出问题:%s\r\n", localFilePath, errRet.Error())
		}
		return errRet
	}
}
