package fs

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ fs.Handle = (*File)(nil)
var _ fs.HandleReader = (*File)(nil)
var _ fs.HandleReadAller = (*File)(nil)
var _ fs.HandleWriter = (*File)(nil)

var _ fs.HandleFlusher = (*File)(nil)
var _ fs.HandleReleaser = (*File)(nil)

func (fh *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	fh.log().Debugf("Read: %+v", req)

	val, err := fh.getData(fh.inode)
	if err != nil {
		return err
	}

	resp.Data = val

	fh.log().Debugf("Read: data length %v", len(resp.Data))
	return nil
}

// ReadAll .
func (fh *File) ReadAll(ctx context.Context) ([]byte, error) {
	val, err := fh.getData(fh.inode)
	if err != nil {
		return nil, err
	}
	fh.log().Debugf("ReadAll: data length %v", len(val))
	return val, nil
}

// Write to the file handle
func (fh *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	fh.log().Debugf("Write: offset. %v req. %+v", req.Offset, req)

	fh.writeData(fh.inode, req.Data)

	attr, _ := fh.getMetadata(fh.inode)
	resp.Size = len(req.Data)
	attr.Size = uint64(resp.Size)

	if err := fh.putMetadata(attr); err != nil {
		fh.log(err).Errorf("Write: putMetadata failed.")
		return err
	}

	fh.log().Debugf("Write: data length %v", len(req.Data))

	return nil
}

// Release .
func (fh *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	fh.log().Debugf("Release: %+v", req)
	return nil
}

// Flush - experimenting with uploading at flush, this slows operations down till it has been
// completely flushed
func (fh *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	fh.log().Debugf("Flush: %+v.", req)
	return nil
}
