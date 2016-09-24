package main

import (
	"fmt"
	"encoding/hex"
//	"strings"
	"sync"
	"strconv"
	"time"
	"os"
	"github.com/mattn/go-gtk/gtk"
//	"github.com/coreyshuman/picontrol/serial"
	"github.com/coreyshuman/xbeeapi"
	"github.com/coreyshuman/picontrol/arduinoio"
)

const XbeeInterDelay = 500

// timing variables
var lastReceivedControl time.Time

// robot state variables
var armDevice bool = false
var stabilize bool = false
var autoVoice bool = false
var headControl bool = false
var playSW bool = false
var volume int = 20
// robot telemetry variables
var analog [6]int
var buttons0 int = 0
var buttons1 int = 0
var targetAddress []byte

func main() {
	gtk.Init(nil)
	var wg sync.WaitGroup
	quit := make(chan bool)
	var serialXBEE int = -1
	var err error
	var i int
	
	// bb-8 body address
	targetAddress = []byte{0x00, 0x13, 0xa2, 0x00, 0x40, 0x90, 0x29, 0x23}
	
	for i=0; i<6; i++ {
		analog[i] = 0
	}
	
	// Initialize APIs
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
		xbeeapi.Close() // exit xbee api
		return
	}
	
	// configure xbee api and start job
	xbeeapi.SetupErrorHandler(errorCallback)
	xbeeapi.SetupModemStatusCallback(modemStatusCallback)
	xbeeapi.SetupATCommandCallback(atCommandCallback)
	xbeeapi.SetupReceivePacketCallback(receivePacketCallback)
	xbeeapi.Begin()
	fmt.Println("XBEE: " + fmt.Sprintf("%d",serialXBEE))
	
	// configure arduino api and start job
	arduinoio.SetupErrorHandler(errorCallback)
	arduinoio.AddHandler(0x10, aioGetAllCB)
	arduinoio.Begin()
	fmt.Println("ARDUINO: " + fmt.Sprintf("%d",serialAIO))

	// Initialize GUI
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("Pi Controller")
	window.SetIconName("gtk-dialog-info")
	window.Connect("destroy", func() {
		quit <- true
		wg.Wait()
		gtk.MainQuit()
	})
	
	/*
	---------------
	| Btns | Telem|
	|      |      |
	---------------
	
	*/
	vbox := gtk.NewVBox(false, 1)

	btnArm := gtk.NewButtonWithLabel("Arm Device")
	btnArm.Clicked(func() {
		fmt.Println("button clicked:", btnArm.GetLabel())
		if (armDevice) {
			armDevice = false
			btnArm.SetLabel("Arm Device")
		} else {
			armDevice = true
			btnArm.SetLabel("Disarm Device")
		}
	})
	btnStabilize := gtk.NewButtonWithLabel("Enable IMU")
	btnStabilize.Clicked(func() {
		fmt.Println("button clicked:", btnStabilize.GetLabel())
		if (stabilize) {
			stabilize = false
			btnStabilize.SetLabel("Enable IMU")
		} else {
			stabilize = true
			btnStabilize.SetLabel("Disable IMU")
		}
	})
	btnAutoVoice := gtk.NewButtonWithLabel("Enable Voice")
	btnAutoVoice.Clicked(func() {
		fmt.Println("button clicked:", btnAutoVoice.GetLabel())
		if (autoVoice) {
			autoVoice = false
			btnAutoVoice.SetLabel("Enable Voice")
		} else {
			autoVoice = true
			btnAutoVoice.SetLabel("Disable Voice")
		}
	})
	btnHeadControl := gtk.NewButtonWithLabel("Enable Head Cont.")
	btnHeadControl.Clicked(func() {
		fmt.Println("button clicked:", btnHeadControl.GetLabel())
		if (headControl) {
			headControl = false
			btnHeadControl.SetLabel("Enable Head Cont.")
		} else {
			headControl = true
			btnHeadControl.SetLabel("Disable Head Cont.")
		}
	})
	volBar := gtk.NewProgressBar()
	volBar.SetFraction(float64(volume)/63.0)
	btnVolUp := gtk.NewButtonWithLabel("Vol UP")
	btnVolUp.Clicked(func() {
		fmt.Println("button clicked:", btnVolUp.GetLabel())
		if (volume < 0x3F) {
			volume ++
			volBar.SetFraction(float64(volume)/63.0)
			// send volume command to device
			err = sendVolume(volume)
		}
	})
	btnVolDown := gtk.NewButtonWithLabel("Vol DOWN")
	btnVolDown.Clicked(func() {
		fmt.Println("button clicked:", btnVolDown.GetLabel())
		if (volume > 0x00) {
			volume --
			volBar.SetFraction(float64(volume)/63.0)
			// send volume command to device
			err = sendVolume(volume)
		}
	})
	btnPlaySW := gtk.NewButtonWithLabel("Play SW")
	btnPlaySW.Clicked(func() {
		fmt.Println("button clicked:", btnPlaySW.GetLabel())
		if (playSW) {
			playSW = false
			btnPlaySW.SetLabel("Play SW")
			err = sendStop(1)
		} else {
			playSW = true
			btnPlaySW.SetLabel("Stop SW")
			err = sendPlay(1, "sw01")
		}
	})
	
	
	/*
	      0            1            2
	--------------------------
	| Arm       | Volume UP  |      0
	-------------------------- 
	| stable    | Volume Disp|      1
	--------------------------
	| Voice     | Volume Down|      2
	--------------------------
	| Head cont.| Play SW    |      3
	--------------------------
	                                4
	*/
	// column left->right, row top->down
	// column start, column stop, row start, row stop
	btnTable := gtk.NewTable(2,4,true)
	// column 1
	btnTable.Attach(btnArm, 0, 1, 0, 1, gtk.FILL, gtk.FILL, 5, 5)
	btnTable.Attach(btnStabilize, 0, 1, 1, 2, gtk.FILL, gtk.FILL, 5, 5)
	btnTable.Attach(btnAutoVoice, 0, 1, 2, 3, gtk.FILL, gtk.FILL, 5, 5)
	btnTable.Attach(btnHeadControl, 0, 1, 3, 4, gtk.FILL, gtk.FILL, 5, 5)
	// column 2
	btnTable.Attach(btnVolUp, 1, 2, 0, 1, gtk.FILL, gtk.FILL, 5, 5)
	btnTable.Attach(volBar, 1, 2, 1, 2, gtk.FILL, gtk.FILL, 5, 5)
	btnTable.Attach(btnVolDown, 1, 2, 2, 3, gtk.FILL, gtk.FILL, 5, 5)
	btnTable.Attach(btnPlaySW, 1, 2, 3, 4, gtk.FILL, gtk.FILL, 5, 5)
	
	vbox.Add(btnTable)
	
	/*
	// show input values as read
	telemTable := gtk.NewTable(4, 7,true)
	// left stick y bar
	lyscale := gtk.NewVScaleWithRange(0, 1024, 1)
	// left stick x bar
	lxscale := gtk.NewHScaleWithRange(0, 1024, 1)
	telemTable.Attach(lyscale, 0,1,0,2,gtk.FILL, gtk.FILL, 5, 5)
	telemTable.Attach(lxscale, 0,3,2,3,gtk.FILL, gtk.FILL, 5, 5)
	// right stick y bar
	ryscale := gtk.NewVScaleWithRange(0, 1024, 1)
	// right stick x bar
	rxscale := gtk.NewHScaleWithRange(0, 1024, 1)
	telemTable.Attach(ryscale, 4,5,0,2,gtk.FILL, gtk.FILL, 5, 5)
	telemTable.Attach(rxscale, 4,7,2,3,gtk.FILL, gtk.FILL, 5, 5)
	
	vbox.Add(telemTable)
	*/
	//--------------------------------------------------------
	// Event
	//--------------------------------------------------------
	window.Add(vbox)
	window.SetSizeRequest(480, 280)
	window.ShowAll()
	
	getXBEEInfo()
	// subroutine to send telemetry
	go func() {
		//var d []byte
		var err error
		wg.Add(1)
		fmt.Println("Begin sending telemetry...")
		for {
			
			select {
			case <- quit:
				closeApp()
				wg.Done()
				return
			default:
				_, _, err = arduinoio.SendGetAllDataCommand()
				if err != nil {
					fmt.Println("Send Command aio error: " + err.Error())
				}
				time.Sleep(time.Millisecond*50)
				err = sendFullTelemetry()
				if err != nil {
					fmt.Println("Send Packet xbee error: " + err.Error())
				}
				time.Sleep(time.Millisecond*50)
				// clear telemetry data if not updated recently
				if(time.Since(lastReceivedControl) > time.Second/2) {
					for i=0; i<6; i++ {
						analog[i] = 0
					}
					buttons0 = 0
					buttons1 = 0
					// todo: put error message on screen
					lastReceivedControl = time.Now()
				}
			}		
		}
	}()
	
	gtk.Main()
}

// todo - algorithms need work
func scale(val float64, min float64, max float64, outMin float64, outMax float64) float64 {
	denom := 1.0
	y := 0.0
	if outMin - min != 0 {
		denom = outMin - min
		y = (outMax - max) / denom * val - min + outMin
	} else {
		y = outMax / max * val - min + outMin
	}
	return y
}

func formatTelemetry() (out []byte) {
	outs := "tel "
	var digital int = 0
	var alg [6]int
	var i int
	var temp int
	
	for i=0; i<6; i++ {
		alg[i] = analog[i]
	}

	// switch left stick x and y
	   temp = alg[2]
	   	alg[2] = alg[3]
		       alg[3] = temp

	if(headControl) {
		alg[4] = alg[2]
		alg[2] = 512
	}
	
	for i=0; i<6; i++ {
		outs += fmt.Sprintf("%04X", alg[i]*0xFFFF/1023) + ","
	}
	outs += "0000,"
	digital = btoi(armDevice) << 15 | btoi(stabilize) << 14 | btoi(autoVoice) << 13
	outs += fmt.Sprintf("%04X", digital)
	outs += "\n";
	fmt.Println(outs)
	
	out = []byte(outs)
	return
}

func btoi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func closeApp() {
	xbeeapi.Close()
	arduinoio.End()
	fmt.Println("Closing...")
	time.Sleep(time.Millisecond*1000)
	fmt.Println("Exit.")
}

func errorCallback(e error) {
	fmt.Println(e.Error())
}

func getXBEEInfo() {
	/******** Get XBEE configuration info *********/
	// get serial number high
	time.Sleep(time.Millisecond*XbeeInterDelay)
	d, n, err := xbeeapi.SendATCommand([]byte{byte('S'), byte('H')}, nil)
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
	
	// get PAN ID
	time.Sleep(time.Millisecond*XbeeInterDelay)
	d, n, err = xbeeapi.SendATCommand([]byte{byte('I'), byte('D')}, nil)
	if(err != nil) {
		fmt.Println("Send AT error: " + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Sent (%d): ", n) + hex.Dump(d))
	}
}


/************** Callback Functions ****************/
var atCommandCallback xbeeapi.ATCommandCallbackFunc = func(frameId byte, data []byte) {
	fmt.Println("AT Response: ")
	fmt.Println(hex.Dump(data))
}

var receivePacketCallback xbeeapi.ReceivePacketCallbackFunc = func(destinationAddress64 [8]byte, destinationAddress16 [2]byte, receiveOptions byte, data []byte) {
	
}

var modemStatusCallback xbeeapi.ModemStatusCallbackFunc = func(status byte) {
	modemStatus := xbeeapi.GetModemStatusDescription(status)
	fmt.Println("Modem Status: " + modemStatus)
}

func aioGetAllCB(d []byte) {
	var i int
	// todo: checksum verification
	fmt.Println("GetAll Callback: ")
    for i = 0; i < 6; i++ {
        analog[i] = (int(d[i*2]) * 256) + int(d[i*2+1])
        s := strconv.Itoa(analog[i])
        fmt.Print(s)
        fmt.Print(", ")
    }
	buttons0 = int(d[i])
	i++
	buttons1 = int(d[i])
    fmt.Println(".")
	lastReceivedControl = time.Now()
}



/************** Send Telemetry/Commands Functions ****************/
func handleSendError(err error) () {
	fmt.Println("Send Error: " + err.Error())
}

func sendFullTelemetry() (error) {
	d := formatTelemetry()
	_, _, err := xbeeapi.SendPacket(targetAddress, nil, 0x00, d)
	return err
}

func sendVolume(volume int) (error) {
	d := []byte{'v', 'o', 'l', ' ', 0x00, '\n'}
	d[4] = byte(volume)
	_, _, err := xbeeapi.SendPacket(targetAddress, nil, 0x00, d)
	return err
}

func sendStop(channel int) (error) {
	d := []byte{'s', 't', 'p', ' ', byte(channel), '\n'}
	_, _, err := xbeeapi.SendPacket(targetAddress, nil, 0x00, d)
	return err
}

func sendPlay(channel int, filename string) (error) {
	d := []byte{'p', 'l', 'y', ' ', byte(channel), ' ' }
	d = append(d[:], filename...)
	d = append(d[:], '\n')
	_, _, err := xbeeapi.SendPacket(targetAddress, nil, 0x00, d)
	return err
}