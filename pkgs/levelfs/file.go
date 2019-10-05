package levelfs

import (
	"os"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type File struct {
	*FS

	// fs.NodeRef
	inode  uint64
	path   string
	isOpen bool
}

var _ fs.Node = (*File)(nil)
var _ fs.FSInodeGenerator = (*File)(nil)
var _ fs.NodeSetattrer = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)
var _ fs.Handle = (*File)(nil)
var _ fs.HandleReader = (*File)(nil)
var _ fs.NodeRemover = (*File)(nil)

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	f.log().Debugf("Attr: %+v", a)
	defer func() {
		f.log().Debugf("Attr: %+v", a)
		f.log().Debugf("set size: %v", a.Size)
	}()
	return f.attr(ctx, a, f.inode)
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	f.log().Debugf("Setattr: %+v", req)
	if f.isOpen {
		return nil
	}
	if req.Mode&^os.ModePerm != 0 {
		return nil
	}
	f.log().Debugf("%+v", req)
	return f.setattr(ctx, req, resp, f.inode)
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	f.log().Debugf("Open: %+v", req)
	// if !req.Flags.IsReadOnly() {
	// 	return nil, fuse.Errno(syscall.EACCES)
	// }
	f.isOpen = true
	return f.Handler(), nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	f.log().Debugf("Read: %+v", req)
	t := time.Now().String()
	fuseutil.HandleRead(req, resp, []byte(t))
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte("hello world."), nil
}

func (f *File) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	f.log().Debugf("Remove: %+v", req)
	return f.remove(ctx, req, f.path)
}

// Handler
func (f *File) Handler() fs.Handle {
	f.log().Debugf("Handler: ")
	return &FileHandle{File: f}
}

func (f *File) log(err ...error) *logrus.Entry {
	fields := logrus.Fields{
		"path":   f.path,
		"inode":  f.inode,
		"module": "fs_file",
		"file":   getLogFilePath(),
	}
	if len(err) > 0 && err[0] != nil {
		fields["error"] = err[0]
	}
	return logrus.WithFields(fields)
}
