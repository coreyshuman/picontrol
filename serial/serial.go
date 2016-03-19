package serial

import (
	"github.com/tarm/serial"
	"time"
	"bufio"
	"bytes"
)

var	s *serial.Port

func Connect() {
	var err error
	// c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 115200, ReadTimeout: time.Millisecond * 30}
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200, ReadTimeout: time.Millisecond * 30}
	s, err = serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
}

func Disconnect() {
	s.Close()
}

func Send(str string) {
	_, err := s.Write([]byte(str))
	
	if err != nil {
		panic(err)
	}
	
}

func Read() string {
	reader := bufio.NewReader(s)
	
	d, err := reader.ReadBytes('\n')
		
		if err != nil {
			return ""
		}
	n := bytes.IndexByte(d, '\n')
	
	return string(d[:n])
}