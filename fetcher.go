package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	pb "gopkg.in/cheggaaa/pb.v1"
)

const (
	FilePrefix string = "gegen-"
)

type HTTPFetcher struct {
	Location   string
	ScratchDir string
}

func (f *HTTPFetcher) MakeURL(fileName string) (string, error) {
	if len(fileName) == 0 {
		return "", MakeError("Empty file name!")
	}
	return strings.Join([]string{f.Location, fileName}, "/"), nil
}

func (f *HTTPFetcher) WriteFile(url string, file *os.File) (int64, error) {
	res, err := http.Get(url)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	size := res.ContentLength

	bar := pb.New(int(size)).SetUnits(pb.U_BYTES)
	bar.Start()

	writer := io.MultiWriter(file, bar)

	written, err := io.Copy(writer, res.Body)
	if err != nil {
		return -1, err
	}
	if written != size {
		return -1, MakeError("Written and actual sizes differ!")
	}
	return written, nil

}

func (f *HTTPFetcher) AcquireFile(filePath string) (bool, error) {
	newName := strings.Join(
		[]string{FilePrefix, path.Base(filePath), "."}, "")

	tempFile, err := ioutil.TempFile(f.ScratchDir, newName)

	if err != nil {
		return false, err
	}
	defer tempFile.Close()

	_, err = f.WriteFile(filePath, tempFile)
	if err != nil {
		return false, err
	}

	return true, nil
}
