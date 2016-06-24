package main

import (
    "github.com/coreyshuman/picontrol/arduinoio"
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
	serialAIO, err := arduinoio.Init(dev, baudn, 1)
	if(err != nil) {
		fmt.Println("Error: " + err.Error())
		return
	}
	arduinoio.SetupErrorHandler(errorCallback)
	arduinoio.AddHandler(0x10, commandResponseCallback)
	arduinoio.Begin()
	fmt.Println("ARDUINO: " + fmt.Sprintf("%d",serialAIO))
    time.Sleep(time.Millisecond*500)
	
    for i := 1; i < 10; i++ {
        // get all data
        d, n, err = arduinoio.SendGetAllDataCommand()
        if(err != nil) {
            fmt.Println("Send ga command error: " + err.Error())
        } else {
            fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
        }
        time.Sleep(time.Millisecond*500)
    }
	
	
	arduinoio.End()
	time.Sleep(time.Millisecond*1000)
	fmt.Println("Exit.")
}

func errorCallback(e error) {
	fmt.Println(e.Error())
}

func commandResponseCallback(d []byte) {
	fmt.Println("Response Callback: ")
	// fmt.Println(hex.Dump(d))
    var analog [6]int
    for i:= 0; i < 6; i++ {
        analog[i] = (int(d[i*2]) * 256) + int(d[i*2+1])
        s := strconv.Itoa(analog[i])
        fmt.Print(s)
        fmt.Print(", ")
    }
    fmt.Println(".")
}