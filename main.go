package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rgbkrk/libvirt-go"
	"github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
)

type EngineConfig struct {
	Name          string
	Owner         string
	Version       float32
	Product       string
	Arch          string
	EngineConfig  map[string]string
	Hypervisors   []map[string]string
	Storages      []map[string]string
	ExtraStorages []map[string]string
}

var bootDevices = [...]string{"hd", "cdrom", "fd", "network"}

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

// Libvirt CPU section
type Host struct {
	UUID string `xml:"uuid"`
	CPU  CPU    `xml:"cpu"`
}

type Caps struct {
	XMLName xml.Name `xml:"capabilities"`
	Host    Host     `xml:"host"`
}

type Feature struct {
	Name string `xml:"name,attr"`
}

type CPU struct {
	Arch          string `xml:"arch,omitempty"`
	Model         string `xml:"model,omitempty"`
	Vendor        string `xml:"vendor,omitempty"`
	ModelFallback string `xml:"mode-fallback,omitempty"`
}

// Virtual Devices section
const vDisk = "disk"
const vNet = "interface"
const vInput = "input"
const vHostdev = "hostdev"
const vSerial = "serial"
const vParallel = "parallel"
const vChannel = "channel"
const vConsole = "console"
const vController = "controller"
const vWatchdog = "watchdog"
const vFilesystem = "filesystem"
const vRedirdev = "redirdev"
const vMemballoon = "memballoon"
const vTPM = "tpm"
const vRNG = "rng"
const vPanic = "panic"

type DiskDriver struct {
	Name  string `xml:"name,attr"`
	Type  string `xml:"type,attr"`
	Cache string `xml:"cache,attr"`
	IO    string `xml:"io,attr"`
}

type DiskSource struct {
	File string `xml:"file,attr"`
}

type DiskTarget struct {
	Dev string `xml:"dev,attr"`
	Bus string `xml:"bus,attr"`
}

type Disk struct {
	Type   string     `xml:"type,attr"`
	Device string     `xml:"device,attr"`
	Driver DiskDriver `xml:"driver"`
	Source DiskSource `xml:"source"`
	Target DiskTarget `xml:"target"`
}

type InterfaceModel struct {
	XMLName xml.Name `xml:"type"`
	Type    string   `xml:"type"`
}

type Mac struct {
	XMLName xml.Name `xml:"mac"`
	Address string   `xml:"address,attr"`
}

type Alias struct {
	XMLName xml.Name `xml:"alias"`
	Name    string   `xml:"name"`
}

type InterfaceSource struct {
	XMLName xml.Name `xml:"source"`
	Network string   `xml:"network,attr"`
	Bridge  string   `xml:"bridge,attr"`
}

type Interface struct {
	XMLName xml.Name `xml:"interface"`
	Type    string   `xml:"type,attr"`
	Mac     Mac
	Source  InterfaceSource
	Model   InterfaceModel
	Alias   Alias
}

type Devices struct {
	Emulator  string `xml:"emulator"`
	Disk      Disk   `xml:"disk"`
	Interface Interface
}

type VCPU struct {
	XMLName   xml.Name `xml:"vcpu"`
	Placement string   `xml:"placement,attr"`
	Size      uint64   `xml:",chardata"`
}

type Clock struct {
	XMLName xml.Name `xml:"clock"`
	Offset  string   `xml:"offset,attr"`
}

type Domain struct {
	XMLName       xml.Name `xml:"domain"`
	Type          string   `xml:"type,attr"`
	Name          string   `xml:"name"`
	UUID          string   `xml:"uuid"`
	Memory        Memory   `xml:"memory"`
	CurrentMemory Memory   `xml:"currentMemory"`
	VCPU          VCPU
	OnPoweroff    string    `xml:"on_poweroff"`
	OnReboot      string    `xml:"on_reboot"`
	OnCrash       string    `xml:"on_crash"`
	OS            OS        `xml:"os,omitempty"`
	CPU           CPU       `xml:"cpu,omitempty"`
	Devices       Devices   `xml:"devices,omitempty"`
	Features      []Feature `xml:"features"`
	Clock         Clock
}

const configuration = "configuration.yaml"

type Memory struct {
	Unit string `xml:"unit,attr"`
	Size int64  `xml:",chardata"`
}

type StoragePermissions struct {
	XMLName xml.Name `xml:"permissions"`
	Owner   string   `xml:"owner"`
	Group   string   `xml:"group"`
	Mode    string   `xml:"mode"`
	Label   string   `xml:"label"`
}

type StorageFormat struct {
	XMLName xml.Name `xml:"format"`
	Type    string   `xml:"type,attr"`
}

type StorageTarget struct {
	XMLName     xml.Name `xml:"target"`
	Path        string   `xml:"path"`
	Format      StorageFormat
	Permissions StoragePermissions
}

type StorageVolume struct {
	XMLName    xml.Name `xml:"volume"`
	Name       string   `xml:"name"`
	Allocation int64    `xml:"allocation"`
	Capacity   Memory   `xml:"capacity"`
	Target     StorageTarget
}

type StoragePool struct {
	XMLName    xml.Name `xml:"pool"`
	Name       string   `xml:"name"`
	UUID       string   `xml:"uuid"`
	Capacity   Memory   `xml:"capacity"`
	Allocation Memory   `xml:"allocation"`
	Available  Memory   `xml:"available"`
	Target     StorageTarget
}

func FancyPrintXML(v interface{}, logfile *os.File) error {
	fmt.Fprintf(logfile, "\n")
	enc := xml.NewEncoder(logfile)
	enc.Indent("DEBUG:\t", "\t")
	if err := enc.Encode(v); err != nil {
		return err
	}
	fmt.Fprintf(logfile, "\n")
	return nil
}

func NewStorageVolume(name string, imgDir string) *StorageVolume {
	path := path.Join(imgDir, name)

	return &StorageVolume{
		Name:       name,
		Allocation: 0,
		Capacity: Memory{
			Unit: "G",
			Size: 10,
		},
		Target: StorageTarget{
			Path: path,
			Format: StorageFormat{
				Type: "qcow2",
			},
			Permissions: StoragePermissions{
				Owner: "107",
				Group: "107",
				Mode:  "0744",
				Label: "virt_image_t",
			},
		},
	}
}

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
				Device: vDisk,
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
				Mac: Mac{
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

const (
	BYTE     uint64 = 1
	KILOBYTE        = 1024 * BYTE
	MEGABYTE        = 1024 * KILOBYTE
	GIGABYTE        = 1024 * MEGABYTE
	NEED            = 50 * GIGABYTE
)

type ThisError struct {
	When time.Time
	What string
}

func (e ThisError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func MakeError(msg string) error {
	return ThisError{When: time.Now(), What: msg}
}

func CheckStorage(p *libvirt.VirStoragePool) error {
	poolInfo, err := p.GetInfo()
	if err != nil {
		return err
	}
	available := poolInfo.GetAvailableInBytes()
	if NEED > available {
		return MakeError("No space available!")
	}
	return nil
}

func AllocateVolume(
	volumeName string,
	imagesPath string,
	pool *libvirt.VirStoragePool) (*StorageVolume, error) {
	storageVolume := NewStorageVolume(volumeName, imagesPath)
	storageXML, err := xml.Marshal(storageVolume)
	if err != nil {
		return nil, err
	}
	_, err = pool.LookupStorageVolByName(volumeName)
	if err != nil {
		fmt.Printf("Allocating volume for %s...\n", volumeName)
		_, err = pool.StorageVolCreateXML(
			string(storageXML),
			libvirt.VIR_STORAGE_VOL_CREATE_PREALLOC_METADATA)
		if err != nil {
			return nil, err
		}
	}
	return storageVolume, nil
}

const STORAGE_POOL_NAME = "default"

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
	for i, _ := range buf {
		temp[0] = buf[i]
		bufs = append(bufs, hex.EncodeToString(temp))

	}
	macString := strings.Join(bufs, ":")
	fmt.Println(macString)
	return macString, nil
}

func main() {
	name := "engine"

	data, err := ioutil.ReadFile(configuration)
	if err != nil {
		panic(err.Error())
	}

	config := EngineConfig{}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err.Error())
	}

	c, err := libvirt.NewVirConnection("qemu:///system")
	if err != nil {
		panic(err.Error())
	}

	capsXml, err := c.GetCapabilities()
	if err != nil {
		panic(err.Error())
	}

	caps := Caps{}

	xml.Unmarshal([]byte(capsXml), &caps)

	virPools, err := c.ListAllStoragePools(libvirt.VIR_STORAGE_POOL_INACTIVE)
	if err != nil {
		panic(err.Error())
	}

	if len(virPools) == 0 {
		panic(MakeError("No pools available!"))
	}

	// Get Storage Pool
	var virPool libvirt.VirStoragePool

	for _, e := range virPools {
		name, err := e.GetName()
		if name == STORAGE_POOL_NAME {
			virPool = e
			break
		}
		if err != nil {
			panic(err.Error())
		}

		isActive, err := e.IsActive()
		if err != nil {
			panic(err.Error())
		}

		if !isActive {
			if err = e.Create(0); err != nil {
				panic(err.Error())
			}
		}
	}

	if err = CheckStorage(&virPool); err != nil {
		panic(err)
	}

	virPoolXML, err := virPool.GetXMLDesc(0)
	if err != nil {
		panic(err.Error())
	}

	storagePool := &StoragePool{}
	if err = xml.Unmarshal([]byte(virPoolXML), storagePool); err != nil {
		panic(err)
	}

	// Create Storage Volume
	imagesPath := storagePool.Target.Path
	storageVolume, err := AllocateVolume(name, imagesPath, &virPool)

	// Create Domain Here
	diskPath := storageVolume.Target.Path
	domain, err := NewDomain(name, diskPath, &caps)
	if err != nil {
		panic(err.Error())
	}

	domainBytes, err := xml.Marshal(&domain)
	if err != nil {
		panic(err.Error())
	}
	domainXML := string(domainBytes)

	fmt.Printf("%v\n", domainXML)

	_, err = c.DomainCreateXML(domainXML, libvirt.VIR_DOMAIN_NONE)
	if err != nil {
		panic(err.Error())
	}

	// Close libvirt connection
	_, err = c.CloseConnection()
	if err != nil {
		panic(err.Error())
	}
}
