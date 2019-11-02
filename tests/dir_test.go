package tests

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.Nil(a.T(), err, "mkdir %s failed, %s %s", md.dir, err, stderr)

		info, err := a.getFileInfo(md.dir)
		a.Logf("get file info aaa, %+v, %s", info, err)
		require.Nil(a.T(), err, err)
		assert.NotNil(a.T(), info)
		assert.True(a.T(), info.IsDir())
		assert.Equal(a.T(), info.Mode().Perm().String(), md.mode.String())

		_, stderr, err = a.doExec("rm", "-rf", md.dir)
		require.Nil(a.T(), err, "rm -rf %s failed, %s %s", md.dir, err, stderr)
	}
}

func (a *AppSuite) TestMkdirAll() {
	for _, md := range []mkdirData{
		mkdirData{"bbb/123", 0755},
		mkdirData{"bbb/zxcv", 0755},
	} {
		_, stderr, err := a.doExec("mkdir", "-p", md.dir)
		require.Nil(a.T(), err, "mkdir %s failed, %s %s", md.dir, err, stderr)

		info, err := a.getFileInfo(md.dir)
		require.Nil(a.T(), err, "get file %s info failed, %s", md.dir, err)
		assert.NotNil(a.T(), info)
		assert.True(a.T(), info.IsDir())
		assert.Equal(a.T(), info.Mode().Perm().String(), md.mode.String())

		_, stderr, err = a.doExec("rm", "-rf", md.dir)
		require.Nil(a.T(), err, "rm -rf %s failed, %s %s", md.dir, err, stderr)
	}
}

func (a *AppSuite) TestMkdirFaild() {
	for _, md := range []mkdirData{
		mkdirData{"ccc/123", 0755},
		mkdirData{"ccc/zxcv", 0755},
	} {
		_, _, err := a.doExec("mkdir", md.dir)
		require.NotNil(a.T(), err, err.Error())
	}
}

func (a *AppSuite) TestMkdirWithMode() {
	// logrus.SetLevel(logrus.DebugLevel)
	// defer logrus.SetLevel(logrus.InfoLevel)
	for _, md := range []mkdirData{
		mkdirData{"d", 0755},
		mkdirData{"dd", 0644},
		mkdirData{"ddd", 0766},
		mkdirData{"dddd", 0744},
		mkdirData{"ddddd", 0666},
		mkdirData{"dddddd", 0444},
		mkdirData{"ddddddd", 0777},
	} {
		_, stderr, err := a.doExec("mkdir", md.dir, "--mode", fmt.Sprintf("%o", md.mode))
		require.Nil(a.T(), err, "mkdir %s failed, %s %s", md.dir, err, stderr)
		info, _ := a.getFileInfo(md.dir)
		require.NotNil(a.T(), info)
		assert.Equal(a.T(), info.Mode().Perm().String(), md.mode.String(), "check mode %s failed", md.mode)
	}
}
