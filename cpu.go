package main

type CPU struct {
	Arch          string `xml:"arch,omitempty"`
	Model         string `xml:"model,omitempty"`
	Vendor        string `xml:"vendor,omitempty"`
	ModelFallback string `xml:"mode-fallback,omitempty"`
}
