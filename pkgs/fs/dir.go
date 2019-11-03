package fs

import (
	"fmt"
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

	dirLogger *logrus.Logger
	inode     uint64
	path      string
}

var _ fs.Node = (*Dir)(nil)
var _ fs.FSInodeGenerator = (*Dir)(nil)
var _ fs.NodeSetattrer = (*Dir)(nil)
var _ fs.NodeStringLookuper = (*Dir)(nil)
var _ fs.HandleReadDirAller = (*Dir)(nil)
var _ fs.NodeMkdirer = (*Dir)(nil)
var _ fs.NodeRemover = (*Dir)(nil)

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	defer d.log().Debugf("Attr: %+v", a.Mode)

	if a.Mode&os.ModePerm != 0 {
		a.Mode = 0755
	}
	if a.Size == 0 {
		a.Size = 4
	}

	inode := d.inode
	if inode == 1 {
		a.Mode = os.ModeDir | a.Mode
		d.log().Debugf("Attr: at root /")
		return nil
	} else if inode == 0 {
		a.Mode = os.ModeDir | a.Mode
		d.log().Debugf("Attr: at zero")
		return nil
	}

	att, err := d.getMetadata(inode)
	if err != nil {
		d.log().Errorf("Attr: getMetadata failed, %s", err)
		return err
	}
	d.log().Debugf("Attr: got metadata: %+v", att)

	a.Inode = inode
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
	d.log().Debugf("Setattr: %s", req)
	return d.setattr(ctx, req, resp, d.inode)
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.log().Debugf("Lookup %+v", name)
	fullpath := filepath.Join(d.path, name)
	inode, err := d.getPath(fullpath)
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
	d.log().Debugf("ReadDirAll: ")
	children, err := d.listChildrenMetadata(d.path)
	if err != nil {
		return nil, err
	}
	dirs := make([]fuse.Dirent, 0, len(children)+2)
	dirs = append(dirs, fuse.Dirent{
		Inode: d.inode,
		Type:  fuse.DT_Dir,
		Name:  ".",
	})
	dirs = append(dirs, fuse.Dirent{
		Inode: d.inode,
		Type:  fuse.DT_Dir,
		Name:  "..",
	})
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
	if req.Mode == 0000 {
		req.Mode = 0755
	}
	var (
		now      = time.Now()
		fullpath = filepath.Join(d.path, req.Name)
		inode    = d.GenerateInode(d.inode, req.Name)
		attr     = &fuse.Attr{
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
	d.log().Debugf("Mkdir: req.mode: %s, attr.mode: %s", req.Mode.String(), attr.Mode.String())

	if err := d.putPath(fullpath, inode); err != nil {
		d.log(err).Errorf("Mkdir: put inode %s failed, %+v", fullpath, inode)
		return nil, err
	}

	if err := d.addChildNode(d.path, req.Name); err != nil {
		d.log(err).Errorf("Mkdir: put children failed, %+v", req)
		return nil, err
	}

	if err := d.putMetadata(attr); err != nil {
		d.log(err).Errorf("Mkdir: put metadata failed, %+v", attr)
		return nil, err
	}
	d.log().Debugf("Mkdir: %s %+v", req.Name, attr)

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

	if err := d.putPath(fullpath, inode); err != nil {
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

// listChildrenMetadata
func (f *FS) listChildrenMetadata(parent string) (map[string]*fuse.Attr, error) {
	children, err := f.getChildren(parent)
	if err != nil {
		return nil, err
	}

	inodes := map[string]uint64{}
	for _, name := range children {
		inode, err := f.getPath(filepath.Join(parent, name))
		if err != nil {
			return nil, err
		}
		inodes[name] = inode
	}

	attrs := map[string]*fuse.Attr{}
	for name, inode := range inodes {
		attr, err := f.getMetadata(inode)
		if err != nil {
			return nil, err
		}
		attrs[name] = attr
	}
	return attrs, nil
}

func (d *Dir) log(err ...error) *logrus.Entry {
	if d.dirLogger == nil {
		d.dirLogger = logrus.New()
		d.dirLogger.Formatter = new(logrus.JSONFormatter)
		// d.dirLogger.SetLevel(logrus.DebugLevel)
	}
	fields := logrus.Fields{
		"path":   d.path,
		"inode":  d.inode,
		"module": "fs_dir",
		"file":   getLogFilePath(),
	}
	if len(err) > 0 && err[0] != nil {
		fields["error"] = err[0].Error()
		fields["error_type"] = fmt.Sprintf("%T", err[0].Error())
	}
	return d.dirLogger.WithFields(fields)
}
