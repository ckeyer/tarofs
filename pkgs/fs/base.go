package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ckeyer/tarofs/pkgs/storage"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

const (
	PrefixINode    = "tarofs_inode_"
	PrefixMetadata = "tarofs_metadata_"
	PrefixPath     = "tarofs_path_"
	PrefixData     = "tarofs_data_"
)

type FS struct {
	metadataStorager storage.MetadataStorager
	dataStorager     storage.DataStorager
	srv              *fs.Server
}

func NewFS(conn *fuse.Conn, ms storage.MetadataStorager, ds storage.DataStorager) *FS {
	return &FS{
		metadataStorager: ms,
		dataStorager:     ds,
		srv:              fs.New(conn, nil),
	}
}

// Serve
func (f *FS) Serve() error {
	logrus.Debug("start levelfs server.")
	return f.srv.Serve(f)
}

var _ fs.FS = (*FS)(nil)
var _ fs.FSInodeGenerator = (*FS)(nil)

func (f *FS) Root() (fs.Node, error) {
	return &Dir{FS: f, path: "/", inode: 1}, nil
}

// name
func (f *FS) GenerateInode(parentInode uint64, name string) uint64 {
	return uint64(time.Now().UnixNano())
}

func (f *FS) setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse, inode uint64) error {
	attr, err := f.getMetadata(inode)
	if err != nil {
		logrus.Errorf("get %v attr failed, %s", inode, err)
		return err
	}

	attr.Mode = req.Mode
	attr.Size = req.Size
	attr.Mtime = time.Now()
	attr.Uid = req.Uid
	attr.Gid = req.Gid

	if err := f.putMetadata(attr); err != nil {
		logrus.Errorf("set attr failed, %s", err)
		return err
	}

	resp.Attr = *attr
	return nil
}

func (f *FS) attr(ctx context.Context, a *fuse.Attr, inode uint64) error {
	if inode == 1 {
		a.Inode = 1
		a.Mode = os.ModeDir | os.ModePerm
		return nil
	}

	a.Inode = inode
	att, err := f.getMetadata(inode)
	if err != nil {
		logrus.Errorf("attr: getMetadata failed, %s", err)
		return err
	}

	a.Size = att.Size
	a.Blocks = att.Blocks
	a.Atime = att.Atime
	a.Mtime = time.Now()
	a.Ctime = att.Ctime
	a.Crtime = att.Crtime
	a.Mode = att.Mode
	a.Nlink = att.Nlink
	a.Uid = att.Uid
	a.Gid = att.Gid
	a.Rdev = att.Rdev
	a.Flags = att.Flags
	a.BlockSize = att.BlockSize

	return nil
}

// remove
func (f *FS) remove(ctx context.Context, req *fuse.RemoveRequest, parent string) error {
	logrus.Debugf("remove file: %+v", req)
	req.Name = filepath.Clean(req.Name)
	children, err := f.getChildren(parent)
	if err != nil {
		logrus.Errorf("remove file, list children faield, %s", children)
		return fuse.ENOENT
	}

	exists := false
	for i, child := range children {
		if child == req.Name {
			exists = true
			children = append(children[:i], children[i+1:]...)
			break
		}
	}
	if !exists {
		return fuse.ENOENT
	}

	if err := f.putChildNode(parent, children); err != nil {
		logrus.Errorf("remove file, reset dir path failed, %s", err)
		return fuse.ENOENT
	}

	fullname := filepath.Join(parent, req.Name)
	inode, err := f.getINode(fullname)
	if err != nil {
		logrus.Errorf("remove file, get inode %s faield, %s", fullname, children)
		return fuse.ENOENT
	}

	f.deleteINode(fullname)
	f.deleteMetadata(inode)

	return nil
}

func (f *FS) getINode(path string) (uint64, error) {
	key := PrefixINode + path
	var inode uint64
	err := f.metadataStorager.Get(key, &inode)
	if err != nil {
		return 0, err
	}
	return inode, nil
}

func (f *FS) putINode(path string, inode uint64) error {
	key := PrefixINode + path
	err := f.metadataStorager.Get(key, nil)
	if err == nil {
		return fuse.EEXIST
	} else if err != errors.ErrNotFound {
		return err
	}

	return f.metadataStorager.Put(key, inode)
}

// deleteINode
func (f *FS) deleteINode(path string) error {
	key := PrefixINode + path
	return f.metadataStorager.Delete(key)
}

// addChildNode
func (f *FS) addChildNode(parent, name string) error {
	oldChildren, err := f.getChildren(parent)
	if err != nil {
		return err
	}

	return f.putChildNode(parent, append(oldChildren, name))
}

func (f *FS) putChildNode(parent string, children []string) error {
	key := PrefixPath + parent
	return f.metadataStorager.Put(key, children)
}

// getChildren
func (f *FS) getChildren(parent string) ([]string, error) {
	key := PrefixPath + parent
	children := []string{}

	if err := f.metadataStorager.Get(key, &children); err != nil {
		if err == errors.ErrNotFound {
			return []string{}, nil
		}
		return nil, err
	}

	return children, nil
}

// putMetadata
func (f *FS) putMetadata(attr *fuse.Attr) error {
	key := PrefixMetadata + fmt.Sprint(attr.Inode)
	return f.metadataStorager.Put(key, attr)
}

// getMetadata
func (f *FS) getMetadata(inode uint64) (*fuse.Attr, error) {
	key := PrefixMetadata + fmt.Sprint(inode)
	attr := &fuse.Attr{}

	if err := f.metadataStorager.Get(key, attr); err != nil {
		return nil, err
	}
	return attr, nil
}

// deleteMetadata
func (f *FS) deleteMetadata(inode uint64) error {
	key := PrefixMetadata + fmt.Sprint(inode)
	return f.metadataStorager.Delete(key)
}

func (f *FS) getData(inode uint64) ([]byte, error) {
	key := PrefixData + fmt.Sprint(inode)
	return f.dataStorager.Bytes(key)
}

func (f *FS) writeData(inode uint64, val []byte) error {
	key := PrefixData + fmt.Sprint(inode)
	return f.dataStorager.PutBytes(key, val)
}
func getLogFilePath() string {
	_, file, line, _ := runtime.Caller(2)
	file = strings.TrimPrefix(file, os.Getenv("GOPATH")+"/src/github.com/ckeyer/tarofs/")
	return fmt.Sprintf("%s:%v", file, line)
}
