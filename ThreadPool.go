package main

import (
	"fmt"
	"os"
	"sync"
)

func (cos *COS) removeGzFile(configurename string, isselfGz, gzfileExist bool, file string) (errRet error) {
	if isselfGz == false && gzfileExist == false {
		errRet = os.Remove(file)
		if errRet != nil {
			cos.log.Error("配置文件%s:删除文件%s失败:%s\r\n", configurename, file, errRet.Error())
		}
	}
	return
}

//线程做的事，具体执行压缩和上传文件
func (cos *COS) worker(configurename string, jobs <-chan string, results chan error, waitn *sync.WaitGroup) {
	defer waitn.Done()
	for j := range jobs {
		file := j
		isselfGz, gzfileExist, gzfile, errRet := cos.Compress(configurename, file)
		if errRet != nil {
			cos.log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
			results <- errRet
			continue
		}
		if isselfGz == false && gzfileExist {
			cos.log.Warn("配置文件%s:%s%s\r\n", configurename, file, "此文件已经有压缩文件了")
			results <- errRet
			continue
		}
		filepath, errRet := cos.extractDir(configurename, gzfile)
		if errRet != nil {
			cos.log.Error("配置文件%s:%s\r\n", configurename, errRet.Error())
			results <- errRet
			errRet = cos.removeGzFile(configurename, isselfGz, gzfileExist, gzfile)
			if errRet != nil {
				results <- errRet
			}
			continue
		}
		errRet = cos.uploadFile(configurename, gzfile, filepath)
		if errRet != nil {
			cos.log.Error("配置文件%s:上传文件%s错误\r\n", configurename, file)
			cos.outlog.Error("配置文件%s:上传文件%s错误\r\n", configurename, file)
			results <- errRet
			errRet = cos.removeGzFile(configurename, isselfGz, gzfileExist, gzfile)
			if errRet != nil {
				results <- errRet
			}
			continue
		}
		cos.outlog.Info("配置文件%s:上传文件%s成功\r\n", configurename, file)
		errRet = cos.removeGzFile(configurename, isselfGz, gzfileExist, gzfile)
		if errRet != nil {
			results <- errRet
		}
		cos.mutex.Lock()
		errRet = cos.recordFile(configurename, file, filepath)
		if errRet != nil {
			cos.log.Error("配置文件%s:记录文件%s错误\r\n", configurename, file)
			results <- errRet
		}
		cos.mutex.Unlock()
		results <- errRet
	}
	return
}

//线程池，设置线程数量和上传文件
func (cos *COS) startwork(configurename string, num int, files []string) (errRet error) {

	jobs := make(chan string, len(files))
	results := make(chan error, len(files))
	var waitn sync.WaitGroup
	for i := 0; i < num; i++ {
		waitn.Add(1)
		go cos.worker(configurename, jobs, results, &waitn)

	}
	for _, v := range files {
		jobs <- v
	}
	close(jobs)
	go func() {
		waitn.Wait()
		close(results)
	}()
	for v := range results {
		if v != nil {
			errRet = fmt.Errorf("程序执行出现问题，具体查看日志")
		}
	}
	if errRet != nil {
		cos.outlog.Error("配置文件%s:startwork中出现问题，具体查看日志\r\n", configurename)
	}
	return

}
