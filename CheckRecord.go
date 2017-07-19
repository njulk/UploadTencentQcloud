package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

//将本地文件路径和远程文件路径写入到记录文件内
func recordFile(localfile string, cosfile string, data map[string]string, recordfile string) (updata map[string]string, errRet error) {
	ishas := isInRecord(localfile, cosfile, data)
	if ishas {
		return data, nil
	} else {
		data[localfile] = cosfile
		result := localfile + "                " + cosfile + "\r\n"
		f, err := os.OpenFile(recordfile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Error("打开记录文件%s失败:%s\r\n", recordfile, err.Error())
			return data, err
		}
		defer f.Close()
		_, errRet = f.WriteString(result)
		if errRet != nil {
			log.Warn("记录文件%s写入记录失败:%s\r\n", recordfile, errRet.Error())
			return data, errRet
		}
		return data, nil
	}

}

//判断记录文件是否已经记录了（本地文件路径和远程路径）
func isInRecord(localfile string, cosfile string, data map[string]string) bool {
	if recordfile, ok := data[localfile]; !ok {
		return false
	} else {
		if strings.Compare(recordfile, cosfile) == 0 {
			return true
		} else {
			return false
		}

	}
}

//从记录文件内获取已经记录的本地文件路径和远程路径的数据
func getRecordData(recordfile string) (map[string]string, error) {
	f, err := os.Open(recordfile)
	if err != nil && os.IsNotExist(err) {
		fc, errRet := os.Create(recordfile)
		defer fc.Close()
		if errRet != nil {
			log.Error("创建记录文件%s失败,%s\r\n", recordfile, err.Error())
			return nil, errRet
		}
		_, errRet = fc.WriteString("localfile" + "                       " + "cosfile" + "\r\n")
		if errRet != nil {
			//f.Close()
			//fmt.Println(errRet.Error())
			log.Warn("记录文件%s记录数据时出错:%s\r\n", recordfile, errRet.Error())
			return nil, errRet
		}
		tmp := map[string]string{"localfile": "cosfile"}
		return tmp, nil
	} else if err != nil {
		log.Error("打开记录文件%s失败:%s\r\n", recordfile, err.Error())
		return nil, err
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	data := make(map[string]string)
	for {
		line, err := rd.ReadString('\n')
		if err != nil && io.EOF == err {
			break
		}
		if err != nil {
			log.Warn("日志提取信息错误:%s\r\n", err.Error())
			break
		}
		y := strings.Fields(line)
		data[y[0]] = y[1]

	}
	return data, nil
}
