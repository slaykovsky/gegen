package src

import (
	"encoding/xml"
	"fmt"
	"path"

	libvirt "github.com/rgbkrk/libvirt-go"
)

const (
	Byte     uint64 = 1
	KibiByte        = 1024 * Byte
	MebiByte        = 1024 * KibiByte
	GibiByte        = 1024 * MebiByte
	// Need is needed to check if there's enough space for setup
	Need = 50 * GibiByte
	// StoragePoolName is the default storage pool name to make volumes on
	StoragePoolName = "default"
)

// StoragePermissions represents permissions section
type StoragePermissions struct {
	XMLName xml.Name `xml:"permissions"`
	Owner   string   `xml:"owner"`
	Group   string   `xml:"group"`
	Mode    string   `xml:"mode"`
	Label   string   `xml:"label"`
}

// StorageFormat represents format section
type StorageFormat struct {
	XMLName xml.Name `xml:"format"`
	Type    string   `xml:"type,attr"`
}

// StorageTarget represents targer section
type StorageTarget struct {
	XMLName     xml.Name `xml:"target"`
	Path        string   `xml:"path"`
	Format      StorageFormat
	Permissions StoragePermissions
}

// StorageVolume represents libvirt storage volume object
type StorageVolume struct {
	XMLName    xml.Name `xml:"volume"`
	Name       string   `xml:"name"`
	Allocation int64    `xml:"allocation"`
	Capacity   Memory   `xml:"capacity"`
	Target     StorageTarget
}

// StoragePool represents libvirt storage volume object
type StoragePool struct {
	XMLName    xml.Name `xml:"pool"`
	Name       string   `xml:"name"`
	UUID       string   `xml:"uuid"`
	Capacity   Memory   `xml:"capacity"`
	Allocation Memory   `xml:"allocation"`
	Available  Memory   `xml:"available"`
	Target     StorageTarget
}

// CheckStorage checks if it's enough space available
func CheckStorage(p *libvirt.VirStoragePool) error {
	poolInfo, err := p.GetInfo()
	if err != nil {
		return err
	}
	available := poolInfo.GetAvailableInBytes()
	if Need > available {
		return MakeError("No space available!")
	}
	return nil
}

// NewStorageVolume makes storage volume object
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

// AllocateVolume allocates volume on storage pool
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
