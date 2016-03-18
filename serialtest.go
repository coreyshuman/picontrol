package main

import (
	"fmt"
	"bufio"
	"github.com/tarm/serial"
	"time"
)


func main() {
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 115200, ReadTimeout: time.Second * 1}
	s, err := serial.OpenPort(c)
	reader := bufio.NewReader(s)
	
	if err != nil {
		panic(err)
	}
	
	for i := 0; i < 50; i++ {
		fmt.Println("write\n")
		_, err = s.Write([]byte("corey\n"))
		
		if err != nil {
			panic(err)
		}
		
		fmt.Println("read\n")
		reply, err := reader.ReadBytes('\n')
		
		if err != nil {
			panic(err)
		}
		
		fmt.Println(reply)
		
		time.Sleep(time.Second*1)
	}
	

	fmt.Println("\nEND\n")
	
	s.Close()
	
}