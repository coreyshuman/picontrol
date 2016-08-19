#!/usr/bin/env bash
echo Starting Pi Control...
DISPLAY=:0.0 /home/pi/devel/src/github.com/coreyshuman/picontrol/picontrol /dev/ttyAMA1 115200 /dev/ttyUSB0 115200
echo Program closed.
read test

