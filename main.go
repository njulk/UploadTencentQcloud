package main

/*import (
	"fmt"
)*/

func main() {

	var configureName string = "config.ini"
	paras, err := getPara(configureName)
	if err != nil {
		return
	}
	var objectcos *COS = new(COS)
	objectcos.appid = paras["appid"]
	objectcos.bucket = paras["bucket"]
	objectcos.secretId = paras["secretId"]
	objectcos.secretKey = paras["secretKey"]
	objectcos.region = paras["region"]
	objectcos.uploadFromlocal(paras["localPath"], true)
}
