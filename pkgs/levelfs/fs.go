package levelfs

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	PrefixData = "tarofs_data_"
)

type FS struct {
	Metadata
	srv *fs.Server
}

func NewFS(conn *fuse.Conn, db *leveldb.DB) *FS {
	return &FS{
		Metadata: NewMetadataSvr(db),
		srv:      fs.New(conn, nil),
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
