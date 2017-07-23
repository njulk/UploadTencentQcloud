package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

//将本地文件路径和远程文件路径写入到记录文件内
func recordFile(configurename string, localfile string, cosfile string, data map[string]string, recordfile string) (updata map[string]string, errRet error) {
	ishas := isInRecord(localfile, cosfile, data)
	if ishas {
		return data, nil
	} else {
		data[localfile] = cosfile
		result := localfile + "                " + cosfile + "\r\n"
		f, err := os.OpenFile(recordfile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Error("配置文件%s:打开记录文件%s失败:%s\r\n", configurename, recordfile, err.Error())
			return data, err
		}
		defer f.Close()
		_, errRet = f.WriteString(result)
		if errRet != nil {
			log.Error("配置文件%s:记录文件%s写入记录失败:%s\r\n", configurename, recordfile, errRet.Error())
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
func getRecordData(configurename string, recordfile string) (map[string]string, error) {
	f, err := os.Open(recordfile)
	if err != nil {
		log.Error("配置文件%s:打开记录文件%s失败:%s\r\n", configurename, recordfile, err.Error())
		err = fmt.Errorf("打开记录文件%s失败:%s", recordfile, err.Error())
		return nil, err
	} else {
		defer f.Close()
		fstat, errstat := os.Stat(recordfile)
		if errstat != nil {
			log.Error("配置文件%s:查询记录文件%s信息有误:%s\r\n", configurename, recordfile, errstat.Error())
			errstat = fmt.Errorf("查询记录文件%s信息有误:%s", recordfile, errstat.Error())
			return nil, errstat
		} else if fstat.Size() == 0 {
			frecord, errstat := os.OpenFile(recordfile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if errstat != nil {
				log.Error("配置文件%s:打开文件%s有误:%s\r\n", configurename, recordfile, errstat.Error())
				errstat = fmt.Errorf("打开文件%s有误:%s", recordfile, errstat.Error())
				return nil, errstat
			} else {
				defer frecord.Close()
				_, errstat = frecord.WriteString("localfile" + "                       " + "cosfile" + "\r\n")
				if errstat != nil {
					errstat = fmt.Errorf("文件%s写入信息时出错:%s", recordfile, errstat.Error())
					log.Error("配置文件%s:文件%s写入信息时出错:%s\r\n", configurename, recordfile, errstat.Error())
					return nil, errstat
				} else {
					tmp := map[string]string{"localfile": "cosfile"}
					return tmp, nil
				}

			}
		}
		rd := bufio.NewReader(f)
		data := make(map[string]string)
		for {
			line, err := rd.ReadString('\n')
			if err != nil && io.EOF == err {
				break
			}
			if err != nil {
				log.Error("配置文件%s:日志提取信息错误:%s\r\n", configurename, err.Error())
				err = fmt.Errorf("日志提取信息错误:%s", err.Error())
				return nil, err
			}
			y := strings.Fields(line)
			data[y[0]] = y[1]
		}
		return data, nil
	}

}
