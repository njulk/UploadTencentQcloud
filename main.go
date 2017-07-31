/*package main

func main() {
	start("config.ini")
}*/

package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("命令行缺少配置文件参数")
		return
	}
	x := os.Args[1:]
	configSize = len(x)
	para, seri, _ := testConfig(x)
	var n sync.WaitGroup
	errResult := make(chan error, 20)
	num := runtime.NumCPU()
	runtime.GOMAXPROCS(num)
	for x := 0; x < len(para); x++ {
		n.Add(1)
		go start(para[x], &n, errResult)
	}
	go func() {
		n.Wait()
		close(errResult)
	}()
	for x := range errResult {
		fmt.Println(x.Error())
	}
	for x := 0; x < len(seri); x++ {
		errRet := start(seri[x], &n, nil)
		if errRet != nil {
			fmt.Println(errRet.Error())
		}
	}

}
