package main

import (
	"os"
	"io"
	"os/exec"
	"bufio"
)

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return err
}

func CopyFile(src, dst string) (err error) {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return MakeError("Can't copy not regular file!")
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return MakeError("File doesn't exist!")
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return MakeError("Can't copy not regular file!")
		}
		if os.SameFile(info, dfi) {
			return nil
		}
	}
	if err = os.Link(src, dst); err == nil {
		return nil
	}
	err = copyFileContents(src, dst)
	return nil
}

func InitrdInject(initrd string, injections []string, tempDir string) (err error) {
	for _, fileName := range injections {
		// Copy fileName to tempDir
		CopyFile(fileName, tempDir)
	}

	pushdCommand := exec.Command("pushd", tempDir)
	popdCommand := exec.Command("popd")
	findCommand := exec.Command("find", ".", "-print0")
	cpioCommand := exec.Command("cpio", "-o", "--null", "-Hnewc", "--quiet")
	gzipCommand := exec.Command("gzip")

	initrdFile, err := os.OpenFile(initrd, os.O_APPEND | os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer initrdFile.Close()

	initrdWriter := bufio.NewWriter(initrdFile)
	defer initrdWriter.Flush()

	gzipCommand.Stdin, _ = cpioCommand.StdoutPipe()
	cpioCommand.Stdin, _ = findCommand.StdoutPipe()
	gzipCommand.Stdout = initrdFile

	pushdCommand.Run()

	err = gzipCommand.Start()
	if err != nil {
		return err
	}

	err = cpioCommand.Run()
	if err != nil {
		return err
	}

	err = findCommand.Run()
	if err != nil {
		return err
	}

	err = gzipCommand.Wait()
	if err != nil {
		return err
	}

	popdCommand.Run()

	return nil
}
