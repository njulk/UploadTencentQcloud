package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func recordFile(localfile string, cosfile string, data map[string]string) (updata map[string]string, errRet error) {
	ishas, _ := isInRecord(localfile, cosfile, data)
	if ishas {
		return data, nil
	} else {
		data[localfile] = cosfile
		result := localfile + "                " + cosfile + "\r\n"
		f, err := os.OpenFile("record.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			fmt.Println(err.Error())
			return data, err
		}
		defer f.Close()
		_, errRet = f.WriteString(result)
		if errRet != nil {
			fmt.Println(errRet.Error())
			return data, errRet
		}
		return data, nil
	}

}

func isInRecord(localfile string, cosfile string, data map[string]string) (bool, error) {
	if recordfile, ok := data[localfile]; !ok {
		return false, nil
	} else {
		//fmt.Println(recordfile, cosfile)
		if strings.Compare(recordfile, cosfile) == 0 {
			return true, nil
		} else {
			return false, nil
		}

	}
}

func getRecordData(recordfile string) (map[string]string, error) {
	f, err := os.Open(recordfile)
	defer f.Close()
	if err != nil {
		//fmt.Println(err.Error())
		//return nil, err
		f, errRet := os.Create("record.txt")
		if errRet != nil {
			fmt.Println(errRet.Error())
			return nil, errRet
		}
		_, errRet = f.WriteString("localfile" + "                       " + "cosfile" + "\r\n")
		if errRet != nil {
			//f.Close()
			//fmt.Println(errRet.Error())
			return nil, errRet
		}
		tmp := map[string]string{"localfile": "cosfile"}
		return tmp, nil
	}
	rd := bufio.NewReader(f)
	data := make(map[string]string)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			//fmt.Println(err.Error())
			break
		}
		y := strings.Fields(line)
		data[y[0]] = y[1]

	}
	return data, nil
}
