package fs

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/sirupsen/logrus"
)

// type File struct {
// 	*File
// }

var _ fs.Handle = (*File)(nil)
var _ fs.HandleReader = (*File)(nil)
var _ fs.HandleFlusher = (*File)(nil)
var _ fs.HandleWriter = (*File)(nil)
var _ fs.HandleReleaser = (*File)(nil)

func (fh *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	fh.hdrlog().Debugf("Read: %+v", req)

	val, err := fh.getData(fh.inode)
	if err != nil {
		return err
	}

	resp.Data = val

	fh.hdrlog().Debugf("Read: %+v", resp)
	return nil
}

// Write to the file handle
func (fh *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	fh.hdrlog().Debugf("Write: offset. %v req. %+v", req.Offset, req)

	fh.writeData(fh.inode, req.Data)

	attr, _ := fh.getMetadata(fh.inode)
	resp.Size = len(req.Data)
	attr.Size = uint64(resp.Size)

	if err := fh.putMetadata(attr); err != nil {
		fh.log(err).Errorf("Write: putMetadata failed.")
		return err
	}

	fh.hdrlog().Debugf("Write: data. %s", req.Data)

	return nil
}

func (fh *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	fh.hdrlog().Debugf("Release: %+v", req)
	return nil
}

// Flush - experimenting with uploading at flush, this slows operations down till it has been
// completely flushed
func (fh *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	fh.hdrlog().Debugf("Flush: %+v, %+v", req, ctx)
	return nil
}

func (fh *File) hdrlog(err ...error) *logrus.Entry {
	fields := logrus.Fields{
		"path":   fh.path,
		"inode":  fh.inode,
		"module": "fs_file_handle",
		"file":   getLogFilePath(),
	}
	if len(err) > 0 && err[0] != nil {
		fields["error"] = err[0]
	}
	return fh.flogger.WithFields(fields)
}
