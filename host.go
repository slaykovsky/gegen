package main

type Host struct {
	UUID string `xml:"uuid"`
	CPU  CPU    `xml:"cpu"`
}
