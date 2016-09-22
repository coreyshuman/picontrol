package main

import (
	"github.com/coreyshuman/xbeeapi"
	"github.com/coreyshuman/picontrol/arduinoio"
	"encoding/hex"
	"fmt"
	"time"
	"os"
	"strconv"
)

const XbeeInterDelay = 500
const AIOInterDelay = 500
const ResponseDelay = 50

var targetAddress []byte
var _tick int64
var serialAIO int = -1
var serialXBEE int = -1

func main() {
	var d []byte
	var n int
	var i int
	var err error
	
	// bb-8 body address
	targetAddress = []byte{0x00, 0x13, 0xa2, 0x00, 0x40, 0x90, 0x29, 0x23}
	
	devx := os.Args[1]
	fmt.Println("XBEE")
	fmt.Println("Dev: " + devx)
	baudx := os.Args[2]
	fmt.Println("Baud: " + baudx)
	baudnx, _ := strconv.Atoi(baudx)
	deva := os.Args[3]
	fmt.Println("ARDUINO")
	fmt.Println("Dev: " + deva)
	bauda := os.Args[4]
	fmt.Println("Baud: " + bauda)
	baudna, _ := strconv.Atoi(bauda)
	serialXBEE, err = xbeeapi.Init(devx, baudnx, 1)
	if(err != nil) {
		fmt.Println("Error: " + err.Error())
		return
	}
	serialAIO, err := arduinoio.Init(deva, baudna, 1)
	if(err != nil) {
		fmt.Println("Error: " + err.Error())
		xbeeapi.End() // exit xbee api
		return
	}
	
	// configure xbee api and start job
	xbeeapi.SetupErrorHandler(errorCallback)
	xbeeapi.AddHandler(0x88, xbeeATComCB)
	xbeeapi.Begin()
	fmt.Println("XBEE: " + fmt.Sprintf("%d",serialXBEE))
	
	// configure arduino api and start job
	arduinoio.SetupErrorHandler(errorCallback)
	arduinoio.AddHandler(0x10, aioGetAllCB)
	arduinoio.Begin()
	fmt.Println("ARDUINO: " + fmt.Sprintf("%d",serialAIO))
	
	/******** Get XBEE configuration info *********/
	// get serial number high
	time.Sleep(time.Millisecond*XbeeInterDelay)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('S'), byte('H')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	
	// get serial number low
	time.Sleep(time.Millisecond*XbeeInterDelay)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('S'), byte('L')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	
	// get 16-bit network address
	time.Sleep(time.Millisecond*XbeeInterDelay)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('M'), byte('Y')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	
	// get channel
	time.Sleep(time.Millisecond*XbeeInterDelay)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('C'), byte('H')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
	
	// get operating PAN ID
	time.Sleep(time.Millisecond*XbeeInterDelay)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('O'), byte('P')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}

        // get operating 16-bit PAN
        time.Sleep(time.Millisecond*XbeeInterDelay)
        d, n, err = xbeeapi.SendATCommand([]byte{byte('O'), byte('I')}, nil)
        if(err != nil) {
                fmt.Println("Send AT error: " + err.Error())
        } else {
                fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
        }

	
	/*************** Enter Main Loop **************/
	// Read Arduino data and send to xbee
	for i = 1; i < 10; i++ {
        // get all data
		fmt.Println(".")
        d, n, err = arduinoio.SendGetAllDataCommand()
        if(err != nil) {
            fmt.Println("Send ga command error: " + err.Error())
        } 
        time.Sleep(time.Millisecond*XbeeInterDelay)
    }
	
	closeApp()

}

func closeApp() {
	xbeeapi.End()
	arduinoio.End()
	fmt.Println("Closing...")
	time.Sleep(time.Millisecond*1000)
	fmt.Println("Exit.")
}

func errorCallback(e error) {
	fmt.Println(e.Error())
}


/************** Callback Functions ****************/

func xbeeATComCB(d []byte) {
	fmt.Println("Response Callback: ")
	fmt.Println(hex.Dump(d))
}

func aioGetAllCB(d []byte) {
	fmt.Println("GetAll Callback: ")
	// fmt.Println(hex.Dump(d))
    var analog [6]int
    for i:= 0; i < 6; i++ {
        analog[i] = (int(d[i*2]) * 256) + int(d[i*2+1])
        s := strconv.Itoa(analog[i])
        fmt.Print(s)
        fmt.Print(", ")
    }
    fmt.Println(".")
	// send to xbee
	if serialXBEE != -1 {
		_, _, err := xbeeapi.SendPacket(targetAddress, nil, 0x00, d)
		if err != nil {
			fmt.Println("Send Packet xbee error: " + err.Error())
		}
	}
}
