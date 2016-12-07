package src

import "encoding/xml"

type OSType struct {
	XMLName xml.Name `xml:"type"`
	Arch    string   `xml:"arch,attr"`
	Machine string   `xml:"machine,attr"`
	Type    string   `xml:",chardata"`
}

type OS struct {
	XMLName    xml.Name `xml:"os"`
	Type       OSType
	Init       string `xml:"init"`
	Kernel     string `xml:"kernel"`
	Initrd     string `xml:"initrd"`
	KernelArgs string `xml:"cmdline"`
}
