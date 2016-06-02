package main

import (
	"github.com/coreyshuman/xbeeapi"
	"encoding/hex"
	"fmt"
	"time"
)



func main() {
	serialXBEE, err := xbeeapi.Init("/dev/ttyAMA0", 115200, 30)
	if(err != nil) {
		fmt.Println("Error: " + err.Error())
		return
	}
	xbeeapi.SetupErrorHandler(errorCallback)
	xbeeapi.AddHandler(0x88, commandResponseCallback)
	xbeeapi.Begin()
	fmt.Println("XBEE: " + fmt.Sprintf("%d",serialXBEE))
	
	
	
	time.Sleep(time.Millisecond*1000)
	
	
	xbeeapi.End()
	time.Sleep(time.Millisecond*1000)
	fmt.Println("Exit.")
}

func errorCallback(e error) {
	fmt.Println(e.Error())
}

func commandResponseCallback(d []byte) {
	fmt.Println("Response Callback: ")
	fmt.Println(hex.Dump(d))
}