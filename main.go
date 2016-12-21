package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rgbkrk/libvirt-go"
	"github.com/slaykovsky/gegen/src"

	"gopkg.in/yaml.v2"
)

// FancyPrintXML prints xml from interface
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

func main() {
	name := "engine"
	configuration := "configuration.yaml"

	data, err := ioutil.ReadFile(configuration)
	if err != nil {
		panic(err.Error())
	}

	config := src.EngineConfig{}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err.Error())
	}

	c, err := libvirt.NewVirConnection("qemu:///system")
	if err != nil {
		panic(err.Error())
	}

	capsXML, err := c.GetCapabilities()
	if err != nil {
		panic(err.Error())
	}

	caps := src.Caps{}

	xml.Unmarshal([]byte(capsXML), &caps)

	virPools, err := c.ListAllStoragePools(libvirt.VIR_STORAGE_POOL_INACTIVE)
	if err != nil {
		panic(err.Error())
	}

	if len(virPools) == 0 {
		panic(src.MakeError("No pools available!"))
	}

	// Get Storage Pool
	var virPool libvirt.VirStoragePool

	for _, e := range virPools {
		name, err := e.GetName()
		if name == src.StoragePoolName {
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

	if err = src.CheckStorage(&virPool); err != nil {
		panic(err.Error())
	}

	virPoolXML, err := virPool.GetXMLDesc(0)
	if err != nil {
		panic(err.Error())
	}

	storagePool := &src.StoragePool{}
	if err = xml.Unmarshal([]byte(virPoolXML), storagePool); err != nil {
		panic(err)
	}

	// Create Storage Volume
	imagesPath := storagePool.Target.Path
	storageVolume, err := src.AllocateVolume(name, imagesPath, &virPool)

	// Create Domain Here
	diskPath := storageVolume.Target.Path
	domain, err := src.NewDomain(name, diskPath, &caps)
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
