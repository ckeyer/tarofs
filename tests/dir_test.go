package tests

import (
	"fmt"
	"os"
)

type mkdirData struct {
	dir  string
	mode os.FileMode
}

func (a *AppSuite) TestMkdirAndRm() {
	for _, md := range []mkdirData{
		mkdirData{"aaa", 0755},
		mkdirData{"./aa", 0755},
		mkdirData{".aaa", 0755},
		mkdirData{"aaaaa/", 0755},
		mkdirData{"./a/", 0755},
		mkdirData{".aaaaa/", 0755},
		mkdirData{"_aaaaa", 0755},
	} {
		_, stderr, err := a.doExec("mkdir", md.dir)
		a.Require().Nil(err, "mkdir %s failed, %s %s", md.dir, err, stderr)

		info, err := a.getFileInfo(md.dir)
		a.T().Logf("get file info aaa, %+v, %s", info, err)

		a.Require().Nil(err, err)
		a.Require().NotNil(info)
		a.Require().True(info.IsDir())
		a.Require().Equal(info.Mode().Perm().String(), md.mode.String())

		_, stderr, err = a.doExec("rm", "-rf", md.dir)
		a.Require().Nil(err, "rm -rf %s failed, %s %s", md.dir, err, stderr)
	}
}

func (a *AppSuite) TestMkdirAll() {
	for _, md := range []mkdirData{
		mkdirData{"bbb/123", 0755},
		mkdirData{"bbb/zxcv", 0755},
	} {
		_, stderr, err := a.doExec("mkdir", "-p", md.dir)
		a.Require().Nil(err, "mkdir %s failed, %s %s", md.dir, err, stderr)

		info, err := a.getFileInfo(md.dir)
		a.Require().Nil(err, "get file %s info failed, %s", md.dir, err)
		a.Require().NotNil(info)
		a.Require().True(info.IsDir())
		a.Require().Equal(md.mode.String(), info.Mode().Perm().String())

		_, stderr, err = a.doExec("rm", "-rf", md.dir)
		a.Require().Nil(err, "rm -rf %s failed, %s %s", md.dir, err, stderr)
	}
}

func (a *AppSuite) TestMkdirFaild() {
	for _, md := range []mkdirData{
		mkdirData{"ccc/123", 0755},
		mkdirData{"ccc/zxcv", 0755},
	} {
		_, _, err := a.doExec("mkdir", md.dir)
		a.Require().NotNil(err, err.Error())
	}
}

func (a *AppSuite) TestMkdirWithMode() {
	for _, md := range []mkdirData{
		mkdirData{"d", 0755},
		mkdirData{"dd", 0644},
		mkdirData{"ddd", 0766},
		mkdirData{"dddd", 0744},
		mkdirData{"ddddd", 0666},
		mkdirData{"dddddd", 0444},
		mkdirData{"ddddddd", 0777},
	} {
		_, stderr, err := a.doExec("mkdir", "-m", fmt.Sprintf("%o", md.mode), md.dir)
		a.Require().Nil(err, "mkdir %s failed, %s %s", md.dir, err, stderr)
		info, _ := a.getFileInfo(md.dir)
		a.Require().NotNil(info)
		a.Require().Equal(info.Mode().Perm().String(), md.mode.String(), "check mode %s failed", md.mode)
	}
}
