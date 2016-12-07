package src

import "encoding/xml"

type Caps struct {
	XMLName xml.Name `xml:"capabilities"`
	Host    Host     `xml:"host"`
}
