package main

import (
	"encoding/xml"

	uuid "github.com/satori/go.uuid"
)

const InstallLocation = "http://fedora.inode.at/releases/24/Server/x86_64/os/"

// Domain represents libvirt guest domain object for XML generation
type Domain struct {
	XMLName       xml.Name `xml:"domain"`
	Type          string   `xml:"type,attr"`
	Name          string   `xml:"name"`
	UUID          string   `xml:"uuid"`
	Memory        Memory   `xml:"memory"`
	CurrentMemory Memory   `xml:"currentMemory"`
	VCPU          VCPU
	OnPoweroff    string `xml:"on_poweroff"`
	OnReboot      string `xml:"on_reboot"`
	OnCrash       string `xml:"on_crash"`
	OS            OS     `xml:"os,omitempty"`
	CPU           CPU    `xml:"cpu,omitempty"`
	Devices       Devices
	Features      []Feature `xml:"features"`
	Clock         Clock
}

// NewDomain creates guest object
func NewDomain(name string, disk string, caps *Caps) (*Domain, error) {
	macAddress, err := GenerateMac()
	if err != nil {
		return nil, err
	}
	return &Domain{
		Type: "kvm",
		Name: name,
		UUID: uuid.NewV4().String(),
		VCPU: VCPU{
			Placement: "static",
			Size:      2,
		},
		CPU: CPU{
			Model:  caps.Host.CPU.Model,
			Vendor: caps.Host.CPU.Vendor,
		},
		CurrentMemory: Memory{
			Unit: "GB",
			Size: 2,
		},
		Memory: Memory{
			Unit: "GB",
			Size: 4,
		},
		OS: OS{
			Type: OSType{
				Arch:    "x86_64",
				Machine: "pc",
				Type:    "hvm",
			},
		},
		Clock: Clock{
			Offset: "utc",
		},
		OnPoweroff: "destroy",
		OnReboot:   "destroy",
		OnCrash:    "destroy",
		Devices: Devices{
			Emulator: "/usr/bin/qemu-kvm",
			Disk: Disk{
				Type:   "file",
				Device: "disk",
				Driver: DiskDriver{
					Name:  "qemu",
					Type:  "qcow2",
					Cache: "none",
					IO:    "native",
				},
				Source: DiskSource{
					File: disk,
				},
				Target: DiskTarget{
					Dev: "vda",
					Bus: "virtio",
				},
			},
			Interface: Interface{
				Type: "network",
				Mac: InterfaceMac{
					Address: macAddress,
				},
				Source: InterfaceSource{
					Network: "default",
					Bridge:  "vibr0",
				},
				Model: InterfaceModel{Type: "virtio"},
				Alias: Alias{Name: "net0"},
			},
		},
	}, nil
}
