package main

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	Location     string = "http://mirror.yandex.ru"
	ScratchDir   string = "/tmp"
	FileName     string = "centos/7/os/x86_64/isolinux/initrd.img"
	FileSize     int64  = 43372552
	TempFileName string = "fetcher_test"
	ExpectedUrl  string = "http://mirror.yandex.ru/centos/7/os/x86_64/isolinux/initrd.img"
)

func TestMakeURL(t *testing.T) {
	fetcher := HTTPFetcher{
		Location:   Location,
		ScratchDir: ScratchDir,
	}

	url, err := fetcher.MakeURL(FileName)
	if err != nil {
		t.Fatal(err)
	}
	if url != ExpectedUrl {
		t.Fatal("Expected url: ", ExpectedUrl,
			" doesn't match actual url: ", url)
	}
}

func TestWriteFile(t *testing.T) {
	fetcher := HTTPFetcher{
		Location:   Location,
		ScratchDir: ScratchDir,
	}

	url, err := fetcher.MakeURL(FileName)
	if err != nil {
		t.Fatal(err)
	}

	tempFile, err := ioutil.TempFile(fetcher.ScratchDir, TempFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	fileSize, err := fetcher.WriteFile(url, tempFile)
	if err != nil {
		t.Fatal(err)
	}
	if fileSize != FileSize {
		t.Fatal("Expected file size: ", FileSize,
			" doesn't match actual file size: ", fileSize)
	}
}
