package main

import (
	"os"
)

//线程做的事，具体执行压缩和上传文件
func (cos *COS) worker(id int, jobs <-chan string, results chan<- string) (errRet error) {
	for j := range jobs {
		file := j
		_, filepath := extractDir(file)
		_, errRet = Compress(file)
		if errRet != nil {
			log.Error("压缩%s错误\r\n", file)
			results <- j
			continue
		}
		gzfile := file + ".gz"
		filepath += ".gz"
		errRet = cos.uploadFile(gzfile, filepath)
		if errRet != nil {
			log.Error("上传文件%s错误\r\n", file)
			results <- j
			continue
		}
		errRet = os.Remove(gzfile)
		if errRet != nil {
			log.Error("删除本地文件%s出错\r\n", gzfile)
		}
		mutex.Lock()
		recordData, errRet = recordFile(file, filepath, recordData, recordTxtName)
		if errRet != nil {
			log.Error("记录文件错误\r\n")
		}
		mutex.Unlock()
		results <- j
	}
	return
}

//线程池，设置线程数量和上传文件
func (cos *COS) startwork(num int, files []string) {
	jobs := make(chan string, len(files))
	results := make(chan string, len(files))
	for i := 0; i < num; i++ {
		go cos.worker(i, jobs, results)
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
