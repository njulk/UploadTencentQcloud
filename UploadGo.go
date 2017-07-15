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

func (cos *COS) smallFileupload(localFilePath string, filePath string) (Resbody string, errRet error) {
	//cosFile.code = -1
	url := cos.generateurl(filePath)
	sign, err := cos.createSignature(filePath, false)
	if err != nil {
		errRet = fmt.Errorf("生成签名失败", err.Error())
		return
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentWriter, err := bodyWriter.CreateFormFile("filecontent", localFilePath)
	if err != nil {
		errRet = fmt.Errorf("内容初始化失败(filecontent)", err.Error())
		return
	}
	localFileReader, err := os.Open(localFilePath)
	if err != nil {
		errRet = fmt.Errorf("本地文件打开失败", err.Error())
		return
	}
	defer localFileReader.Close()
	_, err = io.Copy(contentWriter, localFileReader)
	if err != nil {
		errRet = fmt.Errorf("内容拷贝初始化失败", err.Error())
		return
	}
	if _, err = localFileReader.Seek(0, os.SEEK_SET); err != nil {
		errRet = fmt.Errorf(err.Error())
		return
	}
	shaObject := sha1.New()
	if _, err = io.Copy(shaObject, localFileReader); err != nil {
		errRet = fmt.Errorf("本地文件sha1失败", err.Error())
		return
	}
	filesha := fmt.Sprintf("%x", shaObject.Sum(nil))
	contentWriter, errRet = bodyWriter.CreateFormField("op")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentWriter.Write([]byte("upload"))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}

	contentWriter, errRet = bodyWriter.CreateFormField("sha")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentWriter.Write([]byte(filesha))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentWriter, errRet = bodyWriter.CreateFormField("insertOnly")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentWriter.Write([]byte("0"))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := httppost(url, sign, contentType, &buffer)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	Resbody = string(resp)
	return
}

func (cos *COS) uploadinit(localFilePath string, filePath string) (session string, filesize, slice_size int64, errRet error) {
	url := cos.generateurl(filePath)
	sign, errRet := cos.createSignature(filePath, false)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	fileinfo, errRet := os.Stat(localFilePath)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	filesize = fileinfo.Size()
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, errRet := bodyWriter.CreateFormField("op")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte("upload_slice_init"))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("filesize")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("slice_size")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", cosSliceSize)))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("insertOnly")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", 0)))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := httppost(url, sign, contentType, &buffer)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	js, errRet := simplejson.NewJson(resp)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	code, errRet := js.Get("code").Int()
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	if code != 0 {
		fmt.Println(errRet.Error())
		return
	}
	session, errRet = js.Get("data").Get("session").String()
	if errRet != nil {
		errRet = fmt.Errorf("cos初始化失败，结果为:%s", string(resp))
		return
	}
	slice_size, errRet = js.Get("data").Get("slice_size").Int64()
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	return
}

func (cos *COS) upload_slice_finish(filepath, session string, filesize int64) (Resbody string, errRet error) {
	url := cos.generateurl(filepath)
	sign, errRet := cos.createSignature(filepath, false)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, errRet := bodyWriter.CreateFormField("op")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte("upload_slice_finish"))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("session")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(session))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("filesize")
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := httppost(url, sign, contentType, &buffer)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	Resbody = string(resp)
	return
}

func (cos *COS) upload_slice_data(localFilePath string, filesize int64, filepath, session string, slice_size int64) (isOk bool, errRet error) {
	url := cos.generateurl(filepath)
	sign, errRet := cos.createSignature(filepath, false)
	if errRet != nil {
		fmt.Println("1出问题")
		fmt.Println(errRet.Error())
		return
	}
	localFileReader, errRet := os.Open(localFilePath)
	defer localFileReader.Close()
	if errRet != nil {
		fmt.Println("2出问题")
		fmt.Println(errRet.Error())
		return
	}
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
			fmt.Println("3出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		if readlen == 0 && errRet == io.EOF {
			break
		}
		sliceNowNumber++
		contentwriter, errRet := bodyWriter.CreateFormField("op")
		if errRet != nil {
			fmt.Println("4出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte("upload_slice_data"))
		if errRet != nil {
			fmt.Println("5出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		contentwriter, errRet = bodyWriter.CreateFormField("filecontent")
		if errRet != nil {
			fmt.Println("6出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write(readData[0:readlen])
		if errRet != nil {
			fmt.Println("7出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}

		contentwriter, errRet = bodyWriter.CreateFormField("session")
		if errRet != nil {
			fmt.Println("8出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte(session))
		if errRet != nil {
			fmt.Println("9出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		contentwriter, errRet = bodyWriter.CreateFormField("offset")
		if errRet != nil {
			fmt.Println("10出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", offset)))
		if errRet != nil {
			fmt.Println("11出问题")
			fmt.Println(errRet.Error())
			return false, errRet
		}
		contentType := bodyWriter.FormDataContentType()
		_, errRet = httppost(url, sign, contentType, &buffer)
		if errRet != nil {
			fmt.Println(errRet.Error())
			return false, errRet
		}

	}
	return true, errRet
}

func (cos *COS) largeFileupload(localFilePath string, filepath string) (result string, errRet error) {
	session, filesize, slicesize, err := cos.uploadinit(localFilePath, filepath)
	if err != nil {
		fmt.Println(err.Error())
		errRet = fmt.Errorf("%s", err.Error())
		return
	}
	_, errRet = cos.upload_slice_data(localFilePath, filesize, filepath, session, slicesize)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	result, errRet = cos.upload_slice_finish(filepath, session, filesize)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return
	}
	//fmt.Println(result)
	return
}

func (cos *COS) uploadFile(localFilePath string, filePath string) error {
	fileinfo, errRet := os.Stat(localFilePath)
	if errRet != nil {
		fmt.Println(errRet.Error())
		return errRet
	}
	filesize := fileinfo.Size()
	if filesize <= cosPostMaxSize {
		_, errRet = cos.smallFileupload(localFilePath, filePath)
		return errRet
	} else {
		_, errRet = cos.largeFileupload(localFilePath, filePath)
		return errRet
	}
}
