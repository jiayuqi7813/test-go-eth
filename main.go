package main

import (
	"main.go/cmdclient"
)




func main(){
	c:= client.NewCmdClient("http://localhost:8545","./keystore")
	c.Run()

}
