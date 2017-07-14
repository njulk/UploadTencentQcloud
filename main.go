package main

/*import (
	"fmt"
)*/

func main() {

	var objectcos *COS = new(COS)
	objectcos.appid = ""
	objectcos.bucket = "njulk"
	objectcos.secretId = ""
	objectcos.secretKey = ""
	objectcos.region = "gz"
	//resp := objectcos.createDir("llllkkkkk/ls/ll/")
	//resp, _ := objectcos.smallFileupload("1.png", "llllkkkkk/ls/ll/1.png")
	//_, dirs, _ := objectcos.queryDir("/")

	//fmt.Println(extractDir("/sfd"))
	//objectcos.uploadfile("22.tar", "/22.tar")
	objectcos.uploadFromlocal("/bin/bash", true)
}
