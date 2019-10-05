package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"bazil.org/fuse"
	"github.com/ckeyer/tarofs/pkgs/levelfs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	leveldir, mountpoint string
)

type AppSuite struct {
	suite.Suite

	conn *fuse.Conn
	db   *leveldb.DB
}

func init() {
	batch := time.Now().Format("0102T150405")
	leveldir = filepath.Join(os.TempDir(), batch, "leveldb")
	mountpoint = filepath.Join(os.TempDir(), batch, "taro")
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}

// SetupSuite setup
func (a *AppSuite) SetupSuite() {
	for _, path := range []string{leveldir, mountpoint} {
		if err := os.MkdirAll(path, 0755); err != nil {
			logrus.Fatalf("mkdir %s failed, %s", path, err)
		}
	}

	var err error
	a.db, err = leveldb.OpenFile(leveldir, nil)
	if err != nil {
		logrus.Fatalf("open leveldb failed, %s", err)
	}
	a.logf("open leveldb at %s successful.", leveldir)

	a.conn, err = levelfs.Mount(mountpoint)
	if err != nil {
		logrus.Fatal("mount falied, ", err)
	}
	a.logf("mount %s successful.", mountpoint)

	if p := a.conn.Protocol(); !p.HasInvalidate() {
		logrus.Fatalf("kernel FUSE support is too old to have invalidations: version %v", p)
	}

	filesys := levelfs.NewFS(a.conn, a.db)
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
	levelfs.Umount(mountpoint)
	a.conn.Close()
	a.db.Close()
	// os.RemoveAll(mountpoint)
	// os.RemoveAll(leveldir)
}

func (a *AppSuite) log(args ...interface{}) {
	a.Suite.T().Log(args...)
}
func (a *AppSuite) logf(format string, args ...interface{}) {
	a.Suite.T().Logf(format, args...)
}
func (a *AppSuite) err(args ...interface{}) {
	a.Suite.T().Error(args...)
}
func (a *AppSuite) errf(format string, args ...interface{}) {
	a.Suite.T().Errorf(format, args...)
}
