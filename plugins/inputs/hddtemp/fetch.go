package hddtemp

/*
The MIT License (MIT)

Copyright (c) 2016 Mendelson Gusmão

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"strings"
)

type Disk struct {
	DeviceName  string
	Model       string
	Temperature int32
	Unit        string
	Status      string
}

type hddtemp struct {
}

func newhddtemp() *hddtemp {
	return &hddtemp{}
}

func (h *hddtemp) fetch(address string) ([]Disk, error) {
	var (
		err    error
		conn   net.Conn
		buffer bytes.Buffer
		disks  []Disk
	)

	if conn, err = net.Dial("tcp", address); err != nil {
		return nil, err
	}

	if _, err = io.Copy(&buffer, conn); err != nil {
		return nil, err
	}

	fields := strings.Split(buffer.String(), "|")

	for index := 0; index < len(fields)/5; index++ {
		status := ""
		offset := index * 5
		device := fields[offset+1]
		device = device[strings.LastIndex(device, "/")+1:]

		temperatureField := fields[offset+3]
		temperature, err := strconv.ParseInt(temperatureField, 10, 32)

		if err != nil {
			temperature = 0
			status = temperatureField
		}

		disks = append(disks, Disk{
			DeviceName:  device,
			Model:       fields[offset+2],
			Temperature: int32(temperature),
			Unit:        fields[offset+4],
			Status:      status,
		})
	}

	return disks, nil
}
