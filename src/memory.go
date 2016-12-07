package src

type Memory struct {
	Unit string `xml:"unit,attr"`
	Size int64  `xml:",chardata"`
}
