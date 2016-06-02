package main

import (
	"fmt"
	"strings"
	"sync"
	"strconv"
	"time"
	"github.com/mattn/go-gtk/gtk"
	"github.com/coreyshuman/picontrol/serial"
)

func main() {
	gtk.Init(nil)
	var wg sync.WaitGroup
	quit := make(chan bool)
	var serialUSB int = -1
	var serialXBEE int = -1
	var err error

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("Pi Controller")
	window.SetIconName("gtk-dialog-info")
	window.Connect("destroy", func() {
		serial.Disconnect(serialUSB)
		serial.Disconnect(serialXBEE)
		quit <- true
		wg.Wait()
		gtk.MainQuit()
	})
	
	serial.Init()
	serialUSB, err = serial.Connect("/dev/ttyUSB0", 115200, 30)
	serialXBEE, err = serial.Connect("/dev/ttyAMA0", 115200, 30)
	
	fmt.Println("USB: " + fmt.Sprintf("%d",serialUSB))
	fmt.Println("XBEE: " + fmt.Sprintf("%d",serialXBEE))
	
	vbox := gtk.NewVBox(false, 1)

	btnBackward := gtk.NewButtonWithLabel("Backward")
	btnBackward.Clicked(func() {
		var d string
		fmt.Println("button clicked:", btnBackward.GetLabel())
		serial.Send(serialUSB, "sp 40,0,0\n")
		d, err = serial.Read(serialUSB)
		fmt.Println(d)
		serial.Send(serialXBEE, "navarm 0\n" )
	})
	btnForward := gtk.NewButtonWithLabel("Forward")
	btnForward.Clicked(func() {
		var d string
		fmt.Println("button clicked:", btnForward.GetLabel())
		serial.Send(serialUSB, "sp 0,40,0\n")
		d, err = serial.Read(serialUSB)
		fmt.Println(d)
		serial.Send(serialXBEE, "navarm 1\n" )
	})
	btnLeft := gtk.NewButtonWithLabel("Left")
	btnLeft.Clicked(func() {
		var d string
		fmt.Println("button clicked:", btnLeft.GetLabel())
		serial.Send(serialUSB, "sp 0,0,40\n")
		d, err = serial.Read(serialUSB)
		fmt.Println(d)
		serial.Send(serialXBEE, "navacl 1\n" )
	})
	btnRight := gtk.NewButtonWithLabel("Right")
	btnRight.Clicked(func() {
		var d string
		fmt.Println("button clicked:", btnRight.GetLabel())
		serial.Send(serialUSB, "sp 0,40,40\n")
		d, err= serial.Read(serialUSB)
		fmt.Println(d)
		serial.Send(serialXBEE, "navacl 0\n" )
	})
	
	table := gtk.NewTable(3,4,true)
	table.Attach(btnForward, 1, 2, 0, 1, gtk.FILL, gtk.FILL, 5, 5)
	table.Attach(btnBackward, 1, 2, 2, 3, gtk.FILL, gtk.FILL, 5, 5)
	table.Attach(btnLeft, 0, 1, 1, 2, gtk.FILL, gtk.FILL, 5, 5)
	table.Attach(btnRight, 2, 3, 1, 2, gtk.FILL, gtk.FILL, 5, 5)
	
	vbox.Add(table)
	
	// show input values as read
	inputTable := gtk.NewTable(4, 7,true)
	// left stick y bar
	lyscale := gtk.NewVScaleWithRange(0, 1024, 1)
	// left stick x bar
	lxscale := gtk.NewHScaleWithRange(0, 1024, 1)
	inputTable.Attach(lyscale, 0,1,0,2,gtk.FILL, gtk.FILL, 5, 5)
	inputTable.Attach(lxscale, 0,3,2,3,gtk.FILL, gtk.FILL, 5, 5)
	// right stick y bar
	ryscale := gtk.NewVScaleWithRange(0, 1024, 1)
	// right stick x bar
	rxscale := gtk.NewHScaleWithRange(0, 1024, 1)
	inputTable.Attach(ryscale, 4,5,0,2,gtk.FILL, gtk.FILL, 5, 5)
	inputTable.Attach(rxscale, 4,7,2,3,gtk.FILL, gtk.FILL, 5, 5)
	
	vbox.Add(inputTable)

	//--------------------------------------------------------
	// Event
	//--------------------------------------------------------
	window.Add(vbox)
	window.SetSizeRequest(480, 280)
	window.ShowAll()
	// subroutine to poll controller inputs
	fmt.Println("Setup Go Routine")
	go func() {
		throttle := 0.0
		yaw := 0.0
		pitch := 0.0
		roll := 0.0
		outString := ""
		wg.Add(1)
		fmt.Println("Entering XBEE for loop")
		for {
			time.Sleep(time.Millisecond*100)
			select {
			case <- quit:
				wg.Done()
				return
			default:
				outString = "telnav "
				serial.Send(serialUSB, "ga\n")
				time.Sleep(time.Millisecond*15)
				read, err := serial.Read(serialUSB)
				if err != nil {
					continue
				}
				stripCmd := strings.Split(read, " ")
				if len(stripCmd) < 2 {
					break
				}
				data := strings.Split(stripCmd[1], ",")
				if(len(data) >= 4) {
					for i := 0; i < 4; i++ {
						switch(i) {
							case 0:
								val, _ := strconv.ParseFloat(data[i], 64)
								throttle = scale(val, 0, 1024, 0, 255)
								outString += fmt.Sprintf("%02X", int(throttle)) + ","
								lxscale.SetValue(val)
								break;
							case 1:
								val, _ := strconv.ParseFloat(data[i], 64)
								yaw = scale(val, 0, 1024, 0, 255)
								outString += fmt.Sprintf("%02X", int(yaw)) + ","
								lyscale.SetValue(val)
								break;
							case 2:
								val, _ := strconv.ParseFloat(data[i], 64)
								pitch = scale(val, 0, 1024, 0, 255)
								outString += fmt.Sprintf("%02X", int(pitch)) + ","
								rxscale.SetValue(val)
								break;
							case 3:
								val, _ := strconv.ParseFloat(data[i], 64)
								roll = scale(val, 0, 1024, 0, 255)
								outString += fmt.Sprintf("%02X", int(roll))
								ryscale.SetValue(val)
								break;
						}
					}
					serial.Send(serialXBEE, outString + "\n" )
					time.Sleep(time.Millisecond*15)
					read, err = serial.Read(serialXBEE)
					if err != nil {
						continue
					}	
					fmt.Println(read)
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

