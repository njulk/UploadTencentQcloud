package main

import (
	"os"
)

//线程做的事，具体执行压缩和上传文件
func (cos *COS) worker(configurename string, jobs <-chan string, results chan<- string) {
	for j := range jobs {
		file := j
		gzfile, errRet := Compress(configurename, file)
		if errRet != nil {
			log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
			results <- j
			continue
		}
		filepath, errRet := cos.extractDir(configurename, gzfile)
		if errRet != nil {
			log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
			results <- j
			continue
		}
		errRet = cos.uploadFile(configurename, gzfile, filepath)
		if errRet != nil {
			log.Error("配置文件%s:上传文件%s错误\r\n", configurename, file)
			outlog.Error("配置文件%s:上传文件%s错误\r\n", configurename, file)
			results <- j
			continue
		}
		outlog.Info("配置文件%s:上传文件%s成功\r\n", configurename, file)
		errRet = os.Remove(gzfile)
		if errRet != nil {
			log.Error("配置文件%s:删除本地文件%s出错\r\n", configurename, gzfile)
		}
		mutex.Lock()
		recordData, errRet = recordFile(configurename, file, filepath, recordData, recordTxtName)
		if errRet != nil {
			log.Error("配置文件%s:记录文件%s错误\r\n", configurename, file)
		}
		mutex.Unlock()
		results <- j
	}
	return
}

//线程池，设置线程数量和上传文件
func (cos *COS) startwork(configurename string, num int, files []string) {

	jobs := make(chan string, len(files))
	results := make(chan string, len(files))
	for i := 0; i < num; i++ {
		go cos.worker(configurename, jobs, results)
	}
	for _, v := range files {
		jobs <- v
	}
	close(jobs)
	for _, _ = range files {
		<-results
	}
	close(results)
}
