package tests

import (
	"io/ioutil"
	"os"
)

type touchData struct {
	Name string
	Data []byte
}

func (a *AppSuite) TestTouchAndRm() {
	for _, v := range []touchData{
		{Name: "faaa"},
		{Name: ".faaa"},
	} {
		_, stderr, err := a.doExec("touch", v.Name)
		a.Require().Nil(err, "touch file %s failed, %s %s", v.Name, err, stderr)

		info, err := a.getFileInfo(v.Name)
		a.T().Logf("get file info aaa, %+v, %s", info, err)
		a.Require().Nil(err, err)
		a.Require().NotNil(info)
		a.Require().False(info.IsDir())

		_, stderr, err = a.doExec("rm", "-f", v.Name)
		a.Require().Nil(err, "rm -f %s failed, %s %s", v.Name, err, stderr)
	}
}

func (a *AppSuite) TestWriteRead() {
	for _, v := range []touchData{
		{a.absPath("wr_aaa"), []byte("asdfasfasdfsdfasxvaf\nasdf")},
		{a.absPath(".wr_aaa"), []byte("asdfasfasdfsdfasxvaf\nasdf")},
	} {
		f, err := os.Create(v.Name)
		a.Require().Nilf(err, "create %s failed.", v.Name)

		n, err := f.Write(v.Data)
		a.Require().Nilf(err, "write %s failed.", v.Name)
		a.T().Logf("write %s size %v", v.Name, n)

		err = f.Close()
		a.Require().Nilf(err, "close %s failed.", v.Name)

		fi, err := os.Stat(v.Name)
		a.Require().Nilf(err, "stat %s failed.", v.Name)
		a.Require().Equal(n, int(fi.Size()), "equal file size.")

		rbs, err := ioutil.ReadFile(v.Name)
		a.Require().Nilf(err, "read %s failed.", v.Name)
		a.Require().Equal(string(rbs), string(v.Data), "file data")
	}
}
