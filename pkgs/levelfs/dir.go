package levelfs

import (
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	*FS

	inode uint64
	path  string
}

var _ fs.Node = (*Dir)(nil)
var _ fs.FSInodeGenerator = (*Dir)(nil)
var _ fs.NodeSetattrer = (*Dir)(nil)
var _ fs.NodeStringLookuper = (*Dir)(nil)
var _ fs.HandleReadDirAller = (*Dir)(nil)
var _ fs.NodeMkdirer = (*Dir)(nil)
var _ fs.NodeRemover = (*Dir)(nil)

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	defer d.log().Debugf("Attr: %+v", a)

	inode := a.Inode
	if inode == 1 {
		a.Mode = os.ModeDir | os.ModePerm
		a.Size = 4
		return nil
	} else if inode == 0 {
		a.Mode = os.ModeDir | os.ModePerm
		a.Size = 1024
		return nil
	}

	att, err := d.getMetadata(inode)
	if err != nil {
		logrus.Errorf("attr: getMetadata failed, %s", err)
		return err
	}
	d.log().Debugf("Attr: got metadata: %+v", att)

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

func (d *Dir) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if req.Mode&^os.ModePerm != os.ModeDir {
		return nil
	}
	d.log().Debugf("%s", req)
	return d.setattr(ctx, req, resp, d.inode)
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.log().Debugf("Lookup %+v", name)
	fullpath := filepath.Join(d.path, name)
	inode, err := d.getINode(fullpath)
	if err != nil {
		return nil, fuse.ENOENT
	}

	attr, err := d.getMetadata(inode)
	if err != nil {
		return nil, fuse.ENOENT
	}
	if attr.Mode.IsDir() {
		return &Dir{FS: d.FS, path: fullpath, inode: inode}, nil
	}
	return &File{FS: d.FS, path: fullpath, inode: inode}, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.log().Debugf("ReadDirAll")
	children, err := d.listChildrenMetadata(d.path)
	if err != nil {
		return nil, err
	}
	dirs := make([]fuse.Dirent, 0, len(children))
	for name, attr := range children {
		ftype := fuse.DT_File
		if attr.Mode.IsDir() {
			ftype = fuse.DT_Dir
		} else if !attr.Mode.IsRegular() {
			ftype = fuse.DT_Unknown
		}
		dirs = append(dirs, fuse.Dirent{
			Inode: attr.Inode,
			Type:  ftype,
			Name:  name,
		})
	}

	d.log().Debugf("ReadDirAll, ls all: %+v", dirs)
	return dirs, nil
}

// name
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	req.Name = filepath.Clean(req.Name)
	var (
		now      = time.Now()
		fullpath = filepath.Join(d.path, req.Name)
		inode    = d.GenerateInode(d.inode, req.Name)
		attr     = &fuse.Attr{
			Size:   4,
			Inode:  inode,
			Atime:  now,
			Mtime:  now,
			Ctime:  now,
			Crtime: now,
			Mode:   req.Mode | os.ModeDir,
			Nlink:  1,
			Gid:    req.Gid,
			Uid:    req.Uid,
		}
	)
	d.log().Debugf("req.mode: %s, attr.mode: %s", req.Mode.String(), attr.Mode.String())

	if err := d.putINode(fullpath, inode); err != nil {
		d.log(err).Errorf("put inode %s failed, %+v", fullpath, inode)
		return nil, err
	}

	if err := d.addChildNode(d.path, req.Name); err != nil {
		d.log(err).Errorf("put children failed, %+v", req)
		return nil, err
	}

	if err := d.putMetadata(attr); err != nil {
		d.log(err).Errorf("put metadata failed, %+v", attr)
		return nil, err
	}
	d.log().Debugf("Mkdir %s %+v", req.Name, attr)

	return &Dir{FS: d.FS, path: fullpath, inode: inode}, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	var (
		now      = time.Now()
		fullpath = filepath.Join(d.path, req.Name)
		inode    = d.GenerateInode(d.inode, req.Name)
		f        = &File{FS: d.FS, path: fullpath, inode: inode}
		attr     = &fuse.Attr{
			Size:   0,
			Inode:  inode,
			Atime:  now,
			Mtime:  now,
			Ctime:  now,
			Crtime: now,
			Mode:   req.Mode,
			Nlink:  1,
			Gid:    req.Gid,
			Uid:    req.Uid,
		}
	)
	d.log().Debugf("Create %s request: %+v", req.Name, req)
	d.log().Debugf("Create %s attr: %+v", req.Name, attr)

	if err := d.putINode(fullpath, inode); err != nil {
		d.log(err).Errorf("put inode %s failed, %+v", fullpath, inode)
		return f, f, err
	}

	if err := d.addChildNode(d.path, req.Name); err != nil {
		d.log(err).Errorf("put children failed, %+v", req)
		return f, f, err
	}
	d.log().Debugf("create file mode: %+v, %+v", req.Mode, attr.Mode)
	if err := d.putMetadata(attr); err != nil {
		d.log(err).Errorf("put metadata failed, %+v", attr)
		return f, f, err
	}
	d.log().Debugf("put %s metadata %+v", req.Name, attr)

	resp.LookupResponse.Attr = *attr
	resp.LookupResponse.Node = fuse.NodeID(inode)
	resp.LookupResponse.EntryValid = time.Minute * 5
	resp.OpenResponse.Flags = fuse.OpenDirectIO

	return f, f, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	return d.remove(ctx, req, d.path)
}

func (d *Dir) log(err ...error) *logrus.Entry {
	fields := logrus.Fields{
		"path":   d.path,
		"inode":  d.inode,
		"module": "fs_dir",
		"file":   getLogFilePath(),
	}
	if len(err) > 0 && err[0] != nil {
		fields["error"] = err[0]
	}
	return logrus.WithFields(fields)
}
