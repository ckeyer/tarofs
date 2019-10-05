package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

func doExec(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = mountpoint
	buferr, bufout := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stderr = buferr
	cmd.Stdout = bufout
	err := cmd.Run()
	return bufout.String(), buferr.String(), err
}

func getFileInfo(name string) (os.FileInfo, error) {
	return os.Stat(absPath(name))
}

func absPath(name string) string {
	return filepath.Join(mountpoint, name)
}
