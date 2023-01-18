// Copyright 2016 Michael Stapelberg and contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package usb is a minimal, device-specific library which uses
// gousb to communicate with a Fujitsu ScanSnap iX500 via USB.
package usb

import (
	"fmt"
	"sync"

	"github.com/google/gousb"
)

type Device struct {
	dev    *gousb.Device
	iface  *gousb.Interface
	mu     sync.Mutex
	readm  sync.Mutex
	writem sync.Mutex
	input  *gousb.InEndpoint
	output *gousb.OutEndpoint
	done   func()
}

// TODO: move UsbdevfsBulkTransfer and USBDEVFS_* constants to x/sys/unix
type usbdevfsBulkTransfer struct {
	Ep        uint32
	Len       uint32
	Timeout   uint32
	Pad_cgo_0 [4]byte
	Data      *byte
}

const (
	uSBDEVFS_BULK             = 0xc0185502
	uSBDEVFS_CLAIMINTERFACE   = 0x8004550f
	uSBDEVFS_RELEASEINTERFACE = 0x80045510
)

// Constants specific to the Fujitsu ScanSnap iX500
const (
	// product is the USB product id for the ScanSnap iX500
	product = 0x132b

	// vendor is Fujitsu’s USB vendor ID
	vendor = 0x04c5

	// deviceToHost is the USB endpoint used to transfer data from the
	// device to the host
	deviceToHost = 129

	// hostToDevice is the USB endpoint used to transfer data from the
	// host to the device
	hostToDevice = 2
)

// Read transfers up to len(p) bytes from the device to the host via
// blocking USB bulk transfer.
func (u *Device) Read(p []byte) (n int, err error) {
	u.readm.Lock()
	defer u.readm.Unlock()
	return u.input.Read(p)
}

// Write transfers p from the host to the device via blocking USB bulk
// transfer.
func (u *Device) Write(p []byte) (n int, err error) {
	u.writem.Lock()
	defer u.writem.Unlock()
	return u.output.Write(p)
}

// Close releases all resources associated with the Device. The
// Device must not be used after calling Close.
func (u *Device) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// XXX: assumes the scanner always uses interface number 0
	done := u.done
	u.done = nil
	u.input = nil
	u.output = nil
	done()
	return nil
}

// FindDevice returns a ready-to-use Device object for the Fujitsu
// ScanSnap iX500 or a non-nil error if the scanner is not connected.
func FindDevice() (*Device, error) {
	var ctx *gousb.Context
	var done func()
	var dev *gousb.Device
	var usbif *gousb.Interface
	var input *gousb.InEndpoint
	var output *gousb.OutEndpoint

	ctx = gousb.NewContext()

	dev, err := ctx.OpenDeviceWithVIDPID(vendor, product)
	if err != nil {
		return nil, fmt.Errorf("device with product==%q, vendor==%q not found", product, vendor)
		return nil, err
	}
	err = dev.SetAutoDetach(true)
	if err != nil {
		err = fmt.Errorf("set auto detach kernel driver: %w", err)
		goto handleError
	}

	usbif, done, err = dev.DefaultInterface()
	if err != nil {
		err = fmt.Errorf("get default interface: %w", err)
		goto handleError
	}

	input, err = usbif.InEndpoint(deviceToHost)
	if err != nil {
		err = fmt.Errorf("open InEndpoint: %w", err)
		goto handleError
	}

	output, err = usbif.OutEndpoint(hostToDevice)
	if err != nil {
		err = fmt.Errorf("open OutEndpoint: %w", err)
		goto handleError
	}

	return &Device{
		dev:    dev,
		input:  input,
		output: output,
		done: func() {
			done()
			dev.Close()
			ctx.Close()
		},
	}, nil

handleError:
	if done != nil {
		done()
	}
	if dev != nil {
		dev.Close()
	}
	if ctx != nil {
		ctx.Close()
	}
	return nil, err
}
