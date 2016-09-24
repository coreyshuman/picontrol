#!/usr/bin/env bash
echo Starting Pi Control...
DISPLAY=:0.0 /home/pi/devel/src/github.com/coreyshuman/picontrol/picontrol /dev/ttyS0 38400 /dev/ttyUSB0 115200
echo Program closed.
read test

