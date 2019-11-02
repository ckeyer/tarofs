package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.Nil(a.T(), err, "touch file %s failed, %s %s", v.Name, err, stderr)

		info, err := getFileInfo(v.Name)
		a.logf("get file info aaa, %+v, %s", info, err)
		require.Nil(a.T(), err, err)
		assert.NotNil(a.T(), info)
		assert.False(a.T(), info.IsDir())

		_, stderr, err = a.doExec("rm", "-f", v.Name)
		require.Nil(a.T(), err, "rm -f %s failed, %s %s", v.Name, err, stderr)
	}
}

func (a *AppSuite) TestWriteReadAndRm() {
	for _, v := range []touchData{
		{"faaa", nil},
		{".faaa", nil},
	} {
		_, stderr, err := a.doExec("touch", v.Name)
		require.Nil(a.T(), err, "touch file %s failed, %s %s", v.Name, err, stderr)

		info, err := getFileInfo(v.Name)
		a.logf("get file info aaa, %+v, %s", info, err)
		require.Nil(a.T(), err, err)
		assert.NotNil(a.T(), info)
		assert.False(a.T(), info.IsDir())

		_, stderr, err = a.doExec("rm", "-f", v.Name)
		require.Nil(a.T(), err, "rm -f %s failed, %s %s", v.Name, err, stderr)
	}
}
