package src

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"
)

// TODO: Unify it

// Devices represents devices section in guest domain XML
type Devices struct {
	XMLName   xml.Name `xml:"devices,omitempty"`
	Emulator  string   `xml:"emulator"`
	Disk      Disk     `xml:"disk"`
	Interface Interface
}

// Disk represents disk section in devices
type Disk struct {
	Type   string     `xml:"type,attr"`
	Device string     `xml:"device,attr"`
	Driver DiskDriver `xml:"driver"`
	Source DiskSource `xml:"source"`
	Target DiskTarget `xml:"target"`
}

// DiskDriver represents driver section in disk
type DiskDriver struct {
	Name  string `xml:"name,attr"`
	Type  string `xml:"type,attr"`
	Cache string `xml:"cache,attr"`
	IO    string `xml:"io,attr"`
}

// DiskSource represents source section in disk
type DiskSource struct {
	File string `xml:"file,attr"`
}

// DiskTarget represents target section in disk
type DiskTarget struct {
	Dev string `xml:"dev,attr"`
	Bus string `xml:"bus,attr"`
}

// Alias represents alias section for device
type Alias struct {
	XMLName xml.Name `xml:"alias"`
	Name    string   `xml:"name"`
}

// Interface represents interface section in devices
type Interface struct {
	XMLName xml.Name `xml:"interface"`
	Type    string   `xml:"type,attr"`
	Mac     InterfaceMac
	Source  InterfaceSource
	Model   InterfaceModel
	Alias   Alias
}

// InterfaceMac represents mac section int interface
type InterfaceMac struct {
	XMLName xml.Name `xml:"mac"`
	Address string   `xml:"address,attr"`
}

// InterfaceSource represents source section in interface
type InterfaceSource struct {
	XMLName xml.Name `xml:"source"`
	Network string   `xml:"network,attr"`
	Bridge  string   `xml:"bridge,attr"`
}

// InterfaceModel represents model section in interface
type InterfaceModel struct {
	XMLName xml.Name `xml:"type"`
	Type    string   `xml:"type"`
}

// GenerateMac generates mac address string
func GenerateMac() (string, error) {
	var bufs []string
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	buf[0] = (buf[0] | 2) & 0xfe

	if err != nil {
		return "", err
	}

	temp := make([]byte, 1)
	for i := range buf {
		temp[0] = buf[i]
		bufs = append(bufs, hex.EncodeToString(temp))

	}
	macString := strings.Join(bufs, ":")
	fmt.Println(macString)
	return macString, nil
}
