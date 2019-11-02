package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

func (a AppSuite) doExec(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = a.rootDir
	buferr, bufout := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stderr = buferr
	cmd.Stdout = bufout
	err := cmd.Run()
	return bufout.String(), buferr.String(), err
}

func (a AppSuite) getFileInfo(name string) (os.FileInfo, error) {
	return os.Stat(absPath(name))
}

func (a AppSuite) absPath(name string) string {
	return filepath.Join(a.rootDir, name)
}
