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
func (cos *COS) smallFileupload(configurename string, localFilePath string, filePath string) (Resbody string, errRet error) {
	url := cos.generateurl(filePath)
	sign, err := cos.createSignature(configurename, filePath, false)
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s生成签名失败:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentWriter, err := bodyWriter.CreateFormFile("filecontent", localFilePath)
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s内容初始化失败(filecontent):%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	localFileReader, err := os.Open(localFilePath)
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s本地文件打开失败:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	defer localFileReader.Close()
	_, err = io.Copy(contentWriter, localFileReader)
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s内容拷贝失败:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	if _, err = localFileReader.Seek(0, os.SEEK_SET); err != nil {
		errRet = fmt.Errorf("上传小文件%s:seek起点失败:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	shaObject := sha1.New()
	if _, err = io.Copy(shaObject, localFileReader); err != nil {
		errRet = fmt.Errorf("上传小文件%s本地文件sha1失败", localFilePath, err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	filesha := fmt.Sprintf("%x", shaObject.Sum(nil))
	contentWriter, err = bodyWriter.CreateFormField("op")
	if err != nil {
		errRet = fmt.Errorf("上传小文件创建属性op出现错误:%s", err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	_, err = contentWriter.Write([]byte("upload"))
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s创建属性upload出现错误:%s", localFilePath, err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}

	contentWriter, err = bodyWriter.CreateFormField("sha")
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s创建属性sha出现错误:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	_, err = contentWriter.Write([]byte(filesha))
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s写入sha属性内容出现错误:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	contentWriter, err = bodyWriter.CreateFormField("insertOnly")
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s创建insertOnly属性出现错误:%s", localFilePath, err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	_, err = contentWriter.Write([]byte("0"))
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s写入insertOnly属性内容出现错误:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := cos.httppost(configurename, url, sign, contentType, &buffer)
	if errRet != nil {
		errRet = fmt.Errorf("上传小文件%shttp发送时候出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}

	js, err := simplejson.NewJson(resp)
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s时解析响应包json格式失败:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	code, err := js.Get("code").Int()
	if err != nil {
		errRet = fmt.Errorf("上传小文件%s时解析响应包(code)失败:%s", localFilePath, err.Error())
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return
	}
	if 0 != code {
		errRet = fmt.Errorf("上传小文件%s时响应值显示失败：%s", localFilePath, string(resp))
		cos.log.Error("配置文件:%s,%s", configurename, errRet.Error())
		return string(resp), errRet
	}
	Resbody = string(resp)
	return
}

//上传大文件到对应的cos的初始化
func (cos *COS) uploadinit(configurename string, localFilePath string, filePath string) (session string, filesize, slice_size int64, errRet error) {
	url := cos.generateurl(filePath)
	sign, errRet := cos.createSignature(configurename, filePath, false)
	if errRet != nil {
		cos.log.Error("配置文件%S:上传大文件%s初始化创建签名时候出错:%s", configurename, localFilePath, errRet.Error())
		errRet = fmt.Errorf("上传大文件%s初始化创建签名时候出错:%s", localFilePath, errRet.Error())
		return
	}
	fileinfo, errRet := os.Stat(localFilePath)
	if errRet != nil {
		errRet = fmt.Errorf("查询大文件%s信息失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	filesize = fileinfo.Size()
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, errRet := bodyWriter.CreateFormField("op")
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody创建属性op出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte("upload_slice_init"))
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody写入属性op出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("filesize")
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody创建属性filesize出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody写入属性filesize出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("slice_size")
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody创建属slice_size出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", cosSliceSize)))
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody写入属slice_size出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("insertOnly")
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody创建属性insertOnly出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", 0)))
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中httpbody写入属性insertOnly出错:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := cos.httppost(configurename, url, sign, contentType, &buffer)
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中http上传大文件初始化发送失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	js, errRet := simplejson.NewJson(resp)
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中http大文件初始化响应数据json格式初始化失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	code, errRet := js.Get("code").Int()
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中响应数据获得code码失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	if code != 0 {
		errRet = fmt.Errorf("上传文件%s中响应数据获得code码不为0,获得数据失败", localFilePath)
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	session, errRet = js.Get("data").Get("session").String()
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中响应数据获得session码失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	slice_size, errRet = js.Get("data").Get("slice_size").Int64()
	if errRet != nil {
		errRet = fmt.Errorf("上传文件%s中响应数据获得slice_size失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件:%s:%s", configurename, errRet.Error())
		return
	}
	return
}

//上传大文件的尾部到cos
func (cos *COS) upload_slice_finish(configurename, localFilePath, filepath, session string, filesize int64) (Resbody string, errRet error) {
	url := cos.generateurl(filepath)
	sign, errRet := cos.createSignature(configurename, filepath, false)
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部创建签名失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	var buffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&buffer)
	defer bodyWriter.Close()
	contentwriter, errRet := bodyWriter.CreateFormField("op")
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部httpbody创建op属性失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte("upload_slice_finish"))
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部httpbody写入op属性内容upload_slice_finish失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("session")
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部httpbody创建session属性失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(session))
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部httpbody写入session属性失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	contentwriter, errRet = bodyWriter.CreateFormField("filesize")
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部httpbody创建filesize属性失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", filesize)))
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部httpbody写入filesize属性失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	contentType := bodyWriter.FormDataContentType()
	resp, errRet := cos.httppost(configurename, url, sign, contentType, &buffer)
	if errRet != nil {
		errRet = fmt.Errorf("上传大文件%s尾部失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	js, err := simplejson.NewJson(resp)
	if err != nil {
		errRet = fmt.Errorf("上传大文件%s解析响应包json格式失败:%s", localFilePath, err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	code, err := js.Get("code").Int()
	if err != nil {
		errRet = fmt.Errorf("上传大文件%s解析响应包获取code时候失败:%s", configurename, err.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	if 0 != code {
		errRet = fmt.Errorf("上传大文件%s时候响应包code不为0,上传失败:%s", localFilePath, string(resp))
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	Resbody = string(resp)
	return
}

//上传大文件的具体数据到cos
func (cos *COS) upload_slice_data(configurename, localFilePath string, filesize int64, filepath, session string, slice_size int64) (isOk bool, errRet error) {
	url := cos.generateurl(filepath)
	sign, errRet := cos.createSignature(configurename, filepath, false)
	if errRet != nil {
		errRet = fmt.Errorf("大文件%s内容创建签名失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		isOk = false
		return
	}
	localFileReader, errRet := os.Open(localFilePath)
	if errRet != nil {
		errRet = fmt.Errorf("打开大文件%s失败:%s", localFilePath, errRet.Error())
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		isOk = false
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
			errRet = fmt.Errorf("读取大文件%s内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		if readlen == 0 && errRet == io.EOF {
			break
		}
		sliceNowNumber++
		contentwriter, errRet := bodyWriter.CreateFormField("op")
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:httpbody创建op属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte("upload_slice_data"))
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:httpbody写入upload_slice_data属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		contentwriter, errRet = bodyWriter.CreateFormField("filecontent")
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:创建filecontent属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write(readData[0:readlen])
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:写入filecontent属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}

		contentwriter, errRet = bodyWriter.CreateFormField("session")
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:创建session属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte(session))
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:httpbody写入session属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		contentwriter, errRet = bodyWriter.CreateFormField("offset")
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:httpbody创建offset属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		_, errRet = contentwriter.Write([]byte(fmt.Sprintf("%d", offset)))
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:httpbody写入offset属性内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		contentType := bodyWriter.FormDataContentType()
		resp, errRet := cos.httppost(configurename, url, sign, contentType, &buffer)
		if errRet != nil {
			errRet = fmt.Errorf("大文件%s:http发送大文件内容失败:%s", localFilePath, errRet.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		js, err := simplejson.NewJson(resp)
		if err != nil {
			errRet = fmt.Errorf("大文件%s上传时候响应包json解析失败:%s", localFilePath, err.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		code, err := js.Get("code").Int()
		if err != nil {
			errRet = fmt.Errorf("大文件%s上传过程中响应包json解析获取code值时候失败:%s", localFilePath, err.Error())
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}
		if 0 != code {
			errRet = fmt.Errorf("大文件%s上传过程中响应包获得code不为0，上传失败:%s", localFilePath, string(resp))
			cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
			return false, errRet
		}

	}
	return true, errRet
}

//上传大文件的总的综合函数
func (cos *COS) largeFileupload(configurename, localFilePath string, filepath string) (result string, errRet error) {
	session, filesize, slicesize, err := cos.uploadinit(configurename, localFilePath, filepath)
	if err != nil {
		cos.log.Error("配置文件%s:%s", configurename, err.Error())
		errRet = fmt.Errorf("%s", err.Error())
		return
	}
	_, errRet = cos.upload_slice_data(configurename, localFilePath, filesize, filepath, session, slicesize)
	if errRet != nil {
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		return
	}
	result, errRet = cos.upload_slice_finish(configurename, localFilePath, filepath, session, filesize)
	if errRet != nil {
		cos.log.Error("配置文件%s,%s", configurename, errRet.Error())
		return
	}
	return
}

//上传大小文件的综合函数
func (cos *COS) uploadFile(configurename, localFilePath string, filePath string) error {
	fileinfo, errRet := os.Stat(localFilePath)
	if errRet != nil {
		cos.log.Error("配置文件%s:%s", configurename, errRet.Error())
		errRet = fmt.Errorf("%s", errRet.Error())
		return errRet
	}
	filesize := fileinfo.Size()
	if filesize <= cosPostMaxSize {
		_, errRet = cos.smallFileupload(configurename, localFilePath, filePath)
		if errRet != nil {
			cos.log.Error("配置文件%s", errRet.Error())
		}
		return errRet
	} else {
		_, errRet = cos.largeFileupload(configurename, localFilePath, filePath)
		if errRet != nil {
			cos.log.Error("配置文件%s", errRet.Error())
		}
		return errRet
	}
}
