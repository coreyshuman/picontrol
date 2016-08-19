package arduinoio
/* Arduino IO Communication Controller
 * http://github.com/coreyshuman/picontrol/arduino_io
 * (C) 2016 Corey Shuman
 * 6/23/16
 *
 * License: MIT
 */

import (
	"github.com/coreyshuman/serial"
	"github.com/coreyshuman/srbuf"
	//"time"
	"fmt"
	//"bufio"
	//"bytes"
	"errors"
	"container/list"
	"encoding/hex"
)

// receive handler signature
type RxHandlerFunc func([]byte)

// rx handler struct
type RxHandler struct {
	name string
	frameType byte
	handlerFunc func([]byte)
}

var rxHandlerList *list.List
var quit chan bool 

var txBuf *srbuf.SimpleRingBuff
var rxBuf *srbuf.SimpleRingBuff

var errHandler func(error) = nil
var serialAIO int = -1
var err error

var _frameId int = 1

var _running bool = false

////////////////////


func Init(dev string, baud int, timeout int) (int, error) {	
	txBuf = srbuf.Create(256)
	rxBuf = srbuf.Create(256)
	// initialize a serial interface to the xbee module
	serial.Init()
	serialAIO, err = serial.Connect(dev, baud, timeout)
	quit = make(chan bool)
	rxHandlerList = list.New()
	return serialAIO, err
}


func Begin() {
	if serialAIO == -1 {
		return
	}
	
	_running = true
	go func() {
		for {
			select {
			case <- quit:
				break
			default:
				processRxData()
				processTxData()
			}
		}
		// if we get here, dispose and exit
		serial.Disconnect(serialAIO)
	}()
}


func End() {
	if !_running && serialAIO != -1 {
		serial.Disconnect(serialAIO)
	}
	quit <- true
}

// cts todo - avoid repeat for same framdId
func AddHandler(frameType byte, f func([]byte)) {
	var handler RxHandler
	handler.name = "test"
	handler.frameType = frameType
	handler.handlerFunc = f
	rxHandlerList.PushBack(handler)
}

func findHandler(frameType byte) RxHandlerFunc {
	for e := rxHandlerList.Front(); e != nil; e = e.Next() {
		if e.Value.(RxHandler).frameType == frameType {
			return e.Value.(RxHandler).handlerFunc
		}
	}
	return nil
}

func SetupErrorHandler(f func(error)) {
	errHandler = f
}

func processRxData() {
	var ret bool = false
	var frameId byte
	var err error
	var d []byte
	var n int
	
	d = make([]byte, 256)
	n,err = serial.ReadBytes(serialAIO, d)
	// cts todo - improve this
	if err == nil && n > 0 {
		
		for i:=0; i<n; i++ {
			rxBuf.PutByte(d[i])
			//fmt.Println(fmt.Sprintf("Read:[%02X]", d[i]))
		}
	}
	
	for !ret {
		avail := rxBuf.AvailByteCnt()
		if(avail < 8) { // 8 bytes is minimum for complete packet
			break
		}
		p := rxBuf.PeekBytes(3)
		if(p[0] != 0x7E) {
			rxBuf.GetByte() // skip byte, increment buffer
			continue
		}
		n := int(p[1])*256 + int(p[2])
		if(avail < n+4) { // not all data received yet, break for now
			break
		}
		ret = true
		// if we get here, packet is ready to parse
		data := rxBuf.GetBytes(n+4)
		switch(data[3]) { // Frame Type
			case 0x10 : // get all data response
				frameId, data, err = ParseGetAllDataResponse(data)
				break
				
			default:
				err = errors.New("Frame Type not supported: [" + hex.Dump(data[3:4]) + "]")
				break
		}
		if(err != nil) {
			if(errHandler != nil) {
				errHandler(err)
			}
			return
		}
		// fire callback
		handler := findHandler(frameId) 
		if(handler != nil) {
			handler(data)
		} else {
			fmt.Println("No Handler")
		}
	}
}

func processTxData() {
	// send data out of serial (XBEE) port
	for txBuf.AvailByteCnt() > 0 {
		data := txBuf.GetBytes(0)
		serial.SendBytes(serialAIO, data)
	}
}

func CalcChecksum(data []byte)(byte) {
	n := len(data)
	var cs byte = 0

	for i := 0; i < n; i++ {
		cs += data[i]
	}
	return 0xFF - cs
}

/* ***************************************************************
 * SendGetAllDataCommand
 * Sends command to receive all input data from arduino controller
 *
 * ga/n
 * ***************************************************************/
func SendGetAllDataCommand() (d []byte, n int, err error) {
	d = []byte("ga\n")
	n = 3
    err = nil
    
	// cts todo - improve this
	for i := 0; i<len(d); i++ {
		txBuf.PutByte(d[i])
	}

	return
}

/* ***************************************************************
 * ParseGetAllDataResponse
 * Parse a Get All Data response from arduino controller.
 *
 * 0		- Start Delimiter
 * 1-2		- Length
 * 3		- Frame Type (0x10)
 * 4-5  	- Analog chan 1
 * 6-7  	- Analog chan 2
 * 8-9  	- Analog chan 3
 * 10-11  	- Analog chan 4
 * 12-13  	- Analog chan 5
 * 14-15  	- Analog chan 6
 * 16,17	- Digital inputs
 * 18		- checksum
 * ***************************************************************/
func ParseGetAllDataResponse(r []byte) (frameId byte, data []byte, err error) {
	err = nil
	if(r[3] != 0x10) {
		return 0, nil, errors.New("Invalid Frame Type") 
	}
	
	n := int(r[1])*256 + int(r[2])
	
	if(n != len(r) - 4) {
		return 0, nil, errors.New("Frame Length Error: " + fmt.Sprintf("%d, %d", n, len(r)-4)) 
	}
	
	check := CalcChecksum(r[3:n+3])
	if(check != r[n+3]) {
		return 0, nil, errors.New(fmt.Sprintf( "Checksum Error: calc=[%02X] read=[%02X]", check, r[n+3] ) )
	}
	
	// prepare return data
	frameId = r[3]

	data = r[4:18]

	return
}