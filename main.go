package main

import (
	"fmt"
)

func main() {
	err := start()
	if err != nil {
		fmt.Println("执行失败:%s", err.Error())
	}
}
