package main

import "encoding/xml"

type VCPU struct {
	XMLName   xml.Name `xml:"vcpu"`
	Placement string   `xml:"placement,attr"`
	Size      uint64   `xml:",chardata"`
}
