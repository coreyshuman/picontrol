package main

import (
	"github.com/coreyshuman/xbeeapi"
	"encoding/hex"
	"fmt"
	"time"
	"os"
	"strconv"
)



func main() {
	var d []byte
	var n int
	dev := os.Args[1]
	fmt.Println("Dev: " + dev)
	baud := os.Args[2]
	fmt.Println("Baud: " + baud)
	baudn, _ := strconv.Atoi(baud)
	serialXBEE, err := xbeeapi.Init(dev, baudn, 1)
	if(err != nil) {
		fmt.Println("Error: " + err.Error())
		return
	}
	xbeeapi.SetupErrorHandler(errorCallback)
	xbeeapi.AddHandler(0x88, commandResponseCallback)
	xbeeapi.Begin()
	fmt.Println("XBEE: " + fmt.Sprintf("%d",serialXBEE))
	
	// get serial number high
	time.Sleep(time.Millisecond*500)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('S'), byte('H')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	time.Sleep(time.Millisecond*500)
	
	// get serial number low
	time.Sleep(time.Millisecond*500)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('S'), byte('L')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	time.Sleep(time.Millisecond*500)
	
	// get 16-bit network address
	time.Sleep(time.Millisecond*500)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('M'), byte('Y')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	time.Sleep(time.Millisecond*500)
	
	// get channel
	time.Sleep(time.Millisecond*500)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('C'), byte('H')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	time.Sleep(time.Millisecond*500)
	
	// get PAN ID
	time.Sleep(time.Millisecond*500)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('I'), byte('D')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	time.Sleep(time.Millisecond*500)
	
	
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