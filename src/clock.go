package src

import "encoding/xml"

// Clock represent clock offset for guest domain
type Clock struct {
	XMLName xml.Name `xml:"clock"`
	Offset  string   `xml:"offset,attr"`
}
