package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ckeyer/tarofs/pkgs/fs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type AppSuite struct {
	suite.Suite

	fs *fs.FS

	leveldir, rootDir string
}

func TestSuite(t *testing.T) {
	batch := time.Now().Format("0102T150405")
	as := &AppSuite{
		leveldir: filepath.Join(os.TempDir(), batch, "leveldb"),
		rootDir:  filepath.Join(os.TempDir(), batch, "taro"),
	}
	suite.Run(t, as)
}

// SetupSuite setup
func (a *AppSuite) SetupSuite() {
	for _, path := range []string{a.leveldir, a.rootDir} {
		if err := os.MkdirAll(path, 0755); err != nil {
			a.Suite.Failf("SetupSuite Failed", "mkdir %s failed, %s", path, err)
			return
		}
	}

	if p := a.conn.Protocol(); !p.HasInvalidate() {
		logrus.Fatalf("kernel FUSE support is too old to have invalidations: version %v", p)
	}

	filesys := fs.NewFS(a.conn, a.db)
	go func() {
		if err := filesys.Serve(); err != nil {
			logrus.Fatal("start file system serve failed, ", err)
		}
		// Check if the mount process has an error to report.
		<-a.conn.Ready
		if err := a.conn.MountError; err != nil {
			logrus.Fatal("mount file system failed, ", err)
		}
	}()
	time.Sleep(time.Second)
	a.log("start testing.")
}

// TearDownSuite tear down
func (a *AppSuite) TearDownSuite() {
	fs.Umount(rootDir)
	a.conn.Close()
	a.db.Close()
	// os.RemoveAll(rootDir)
	// os.RemoveAll(leveldir)
}

func (a *AppSuite) Log(args ...interface{}) {
	a.Suite.T().Log(args...)
}
func (a *AppSuite) Logf(format string, args ...interface{}) {
	a.Suite.T().Logf(format, args...)
}
func (a *AppSuite) Err(args ...interface{}) {
	a.Suite.T().Error(args...)
}
func (a *AppSuite) Errf(format string, args ...interface{}) {
	a.Suite.T().Errorf(format, args...)
}
