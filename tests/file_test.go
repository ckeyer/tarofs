package tests

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
		a.Assert().NotNil(info)
		a.Assert().False(info.IsDir())

		_, stderr, err = a.doExec("rm", "-f", v.Name)
		a.Require().Nil(err, "rm -f %s failed, %s %s", v.Name, err, stderr)
	}
}

func (a *AppSuite) TestWriteReadAndRm() {
	for _, v := range []touchData{
		{"faaa", nil},
		{".faaa", nil},
	} {
		_, stderr, err := a.doExec("touch", v.Name)
		a.Require().Nil(err, "touch file %s failed, %s %s", v.Name, err, stderr)

		info, err := a.getFileInfo(v.Name)
		a.T().Logf("get file info aaa, %+v, %s", info, err)
		a.Require().Nil(err, err)
		a.Assert().NotNil(info)
		a.Assert().False(info.IsDir())

		_, stderr, err = a.doExec("rm", "-f", v.Name)
		a.Require().Nil(err, "rm -f %s failed, %s %s", v.Name, err, stderr)
	}
}
