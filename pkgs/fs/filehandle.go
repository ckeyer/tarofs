package fs

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/sirupsen/logrus"
)

type FileHandle struct {
	*File
}

var _ fs.Handle = (*FileHandle)(nil)
var _ fs.HandleReader = (*FileHandle)(nil)
var _ fs.HandleFlusher = (*FileHandle)(nil)
var _ fs.HandleWriter = (*FileHandle)(nil)
var _ fs.HandleReleaser = (*FileHandle)(nil)

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	fh.log().Debugf("Read: %+v", req)

	val, err := fh.getData(fh.inode)
	if err != nil {
		return err
	}

	resp.Data = val

	fh.log().Debugf("Read: %+v", resp)
	return nil
}

// Write to the file handle
func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	fh.log().Debugf("Write: %+v", req)

	fh.writeData(fh.inode, req.Data)

	attr, _ := fh.getMetadata(fh.inode)
	attr.Size = uint64(len(req.Data))
	// req.
	fh.putMetadata(attr)
	resp.Size = len(req.Data)
	fh.log().Debugf("Write: %s", req.Data)

	return fh.writeData(fh.inode, req.Data)
}

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	fh.log().Debugf("Release: %+v", req)

	return nil
}

// Flush - experimenting with uploading at flush, this slows operations down till it has been
// completely flushed
func (fh *FileHandle) Flush(ctx context.Context, req *fuse.FlushRequest) error {

	fh.log().Debugf("Flush: %+v", req)
	return nil
}

func (fh *FileHandle) log(err ...error) *logrus.Entry {
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
