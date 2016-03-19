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

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("Pi Controller")
	window.SetIconName("gtk-dialog-info")
	window.Connect("destroy", func() {
		serial.Disconnect()
		quit <- true
		wg.Wait()
		gtk.MainQuit()
	})
	
	serial.Connect()
	
	vbox := gtk.NewVBox(false, 1)

	btnBackward := gtk.NewButtonWithLabel("Backward")
	btnBackward.Clicked(func() {
		fmt.Println("button clicked:", btnBackward.GetLabel())
		serial.Send("sp 40,0,0\n")
		d := serial.Read()
		fmt.Println(d)
	})
	btnForward := gtk.NewButtonWithLabel("Forward")
	btnForward.Clicked(func() {
		fmt.Println("button clicked:", btnForward.GetLabel())
		serial.Send("sp 0,40,0\n")
		d := serial.Read()
		fmt.Println(d)
	})
	btnLeft := gtk.NewButtonWithLabel("Left")
	btnLeft.Clicked(func() {
		fmt.Println("button clicked:", btnLeft.GetLabel())
		serial.Send("sp 0,0,40\n")
		d := serial.Read()
		fmt.Println(d)
	})
	btnRight := gtk.NewButtonWithLabel("Right")
	btnRight.Clicked(func() {
		fmt.Println("button clicked:", btnRight.GetLabel())
		serial.Send("sp 0,40,40\n")
		d := serial.Read()
		fmt.Println(d)
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
	window.SetSizeRequest(300, 300)
	window.ShowAll()
	// subroutine to poll controller inputs
	fmt.Println("Setup Go Routine")
	go func() {
		wg.Add(1)
		for {
			select {
			case <- quit:
				wg.Done()
				return
			default:
				serial.Send("ga\n")
				time.Sleep(time.Millisecond*15)
				read := serial.Read()
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
								lxscale.SetValue(val)
								break;
							case 1:
								val, _ := strconv.ParseFloat(data[i], 64)
								lyscale.SetValue(val)
								break;
							case 2:
								val, _ := strconv.ParseFloat(data[i], 64)
								rxscale.SetValue(val)
								break;
							case 3:
								val, _ := strconv.ParseFloat(data[i], 64)
								ryscale.SetValue(val)
								break;
						}
					}
				}
				time.Sleep(time.Millisecond*100)
			}
		}
	}()
	
	gtk.Main()
	
	
	
}

