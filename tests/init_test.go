package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ckeyer/tarofs/pkgs/fs"
	"github.com/ckeyer/tarofs/pkgs/storage/levelfs"
	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	batch := time.Now().Format("0102T150405")
	as := &AppSuite{
		Suite:    new(suite.Suite),
		leveldir: filepath.Join(os.TempDir(), batch, "leveldb"),
		rootDir:  filepath.Join(os.TempDir(), batch, "taro"),
	}
	suite.Run(t, as)
}

// SetupSuite setup
func (a *AppSuite) SetupSuite() {
	for _, path := range []string{a.leveldir, a.rootDir} {
		if err := os.MkdirAll(path, 0755); err != nil {
			a.Failf("SetupSuite Failed", "mkdir %s failed, %s", path, err)
			return
		}
	}

	stgr, err := levelfs.NewLevelStorage(a.leveldir)
	if err != nil {
		a.Fail("new levelfs storage failed, %s", err)
		return
	}

	a.fs, err = fs.NewFS(a.rootDir, stgr, stgr)
	if err != nil {
		a.Fail("new mount falied, %s", err)
		return
	}
	go func() {
		if err := a.fs.Serve(); err != nil {
			a.Fail("start file system serve failed, ", err)
		}
	}()

	time.Sleep(time.Second)
	go func() {
		select {
		case <-time.Tick(time.Second * 30):
			a.FailNow("timeout. TearDownSuite")
			os.Exit(1)
		}
	}()
	a.T().Logf("root dir: %s", a.rootDir)
	a.T().Log("start testing.")
}

// TearDownSuite tear down
func (a *AppSuite) TearDownSuite() {
	if a.fs != nil {
		a.fs.Close()
	}

	// os.RemoveAll(rootDir)
	// os.RemoveAll(leveldir)
}
