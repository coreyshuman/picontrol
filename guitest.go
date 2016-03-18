package main

import (
	"fmt"
	"github.com/mattn/go-gtk/gtk"
	"github.com/coreyshuman/picontrol/serial"
)

func main() {
	gtk.Init(nil)

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("Pi Controller")
	window.SetIconName("gtk-dialog-info")
	window.Connect("destroy", func() {
		serial.Disconnect()
		gtk.MainQuit()
	})
	
	serial.Connect()

	btnBackward := gtk.NewButtonWithLabel("Backward")
	btnBackward.Clicked(func() {
		fmt.Println("button clicked:", btnBackward.GetLabel())
		serial.Send("b\n")
		d := serial.Read()
		fmt.Println(d)
	})
	btnForward := gtk.NewButtonWithLabel("Forward")
	btnForward.Clicked(func() {
		fmt.Println("button clicked:", btnForward.GetLabel())
		serial.Send("f\n")
		d := serial.Read()
		fmt.Println(d)
	})
	btnLeft := gtk.NewButtonWithLabel("Left")
	btnLeft.Clicked(func() {
		fmt.Println("button clicked:", btnLeft.GetLabel())
		serial.Send("l\n")
		d := serial.Read()
		fmt.Println(d)
	})
	btnRight := gtk.NewButtonWithLabel("Right")
	btnRight.Clicked(func() {
		fmt.Println("button clicked:", btnRight.GetLabel())
		serial.Send("r\n")
		d := serial.Read()
		fmt.Println(d)
	})
	
	table := gtk.NewTable(3,3,true)
	table.Attach(btnForward, 1, 2, 0, 1, gtk.FILL, gtk.FILL, 5, 5)
	table.Attach(btnBackward, 1, 2, 2, 3, gtk.FILL, gtk.FILL, 5, 5)
	table.Attach(btnLeft, 0, 1, 1, 2, gtk.FILL, gtk.FILL, 5, 5)
	table.Attach(btnRight, 2, 3, 1, 2, gtk.FILL, gtk.FILL, 5, 5)
	
	

	//--------------------------------------------------------
	// Event
	//--------------------------------------------------------
	window.Add(table)
	window.SetSizeRequest(300, 150)
	window.ShowAll()
	gtk.Main()
}