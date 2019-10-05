package levelfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"bazil.org/fuse"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

const (
	PrefixINode    = "tarofs_inode_"
	PrefixMetadata = "tarofs_metadata_"
	PrefixPath     = "tarofs_path_"
)

type Metadata struct {
	db *leveldb.DB
}

func NewMetadataSvr(db *leveldb.DB) Metadata {
	return Metadata{db: db}
}

func jsonEncode(val interface{}) []byte {
	bs, _ := json.Marshal(val)
	return bs
}

func jsonDecode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func getLogFilePath() string {
	_, file, line, _ := runtime.Caller(2)
	file = strings.TrimPrefix(file, os.Getenv("GOPATH")+"/src/github.com/ckeyer/tarofs/")
	return fmt.Sprintf("%s:%v", file, line)
}

func (f *Metadata) putINode(path string, inode uint64) error {
	key := PrefixINode + path
	err := f.get(key, nil)
	if err == nil {
		return fuse.EEXIST
	} else if err != errors.ErrNotFound {
		return err
	}

	return f.put(key, inode)
}

func (f *Metadata) getINode(path string) (uint64, error) {
	key := PrefixINode + path
	var inode uint64
	err := f.get(key, &inode)
	if err != nil {
		return 0, err
	}
	return inode, nil
}

// deleteINode
func (f *Metadata) deleteINode(path string) error {
	key := PrefixINode + path
	return f.delete(key)
}

// addChildNode
func (f *Metadata) addChildNode(parent, name string) error {
	oldChildren, err := f.getChildren(parent)
	if err != nil {
		return err
	}

	return f.putChildNode(parent, append(oldChildren, name))
}

func (f *Metadata) putChildNode(parent string, children []string) error {
	key := PrefixPath + parent
	return f.put(key, children)
}

// getChildren
func (f *Metadata) getChildren(parent string) ([]string, error) {
	key := PrefixPath + parent
	children := []string{}

	if err := f.get(key, &children); err != nil {
		if err == errors.ErrNotFound {
			return []string{}, nil
		}
		return nil, err
	}

	return children, nil
}

// putMetadata
func (f *Metadata) putMetadata(attr *fuse.Attr) error {
	key := PrefixMetadata + fmt.Sprint(attr.Inode)
	return f.put(key, attr)
}

// getMetadata
func (f *Metadata) getMetadata(inode uint64) (*fuse.Attr, error) {
	key := PrefixMetadata + fmt.Sprint(inode)
	attr := &fuse.Attr{}

	if err := f.get(key, attr); err != nil {
		return nil, err
	}
	return attr, nil
}

// deleteMetadata
func (f *Metadata) deleteMetadata(inode uint64) error {
	key := PrefixMetadata + fmt.Sprint(inode)
	return f.delete(key)
}

// listChildrenMetadata
func (f *Metadata) listChildrenMetadata(parent string) (map[string]*fuse.Attr, error) {
	children, err := f.getChildren(parent)
	if err != nil {
		return nil, err
	}

	inodes := map[string]uint64{}
	for _, name := range children {
		inode, err := f.getINode(filepath.Join(parent, name))
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

// getData
func (f *Metadata) getData(inode uint64) ([]byte, error) {
	key := PrefixData + fmt.Sprint(inode)
	return f.getValue(key)
}

// writeData
func (f *Metadata) writeData(inode uint64, val []byte) error {
	key := PrefixData + fmt.Sprint(inode)
	return f.putValue(key, val)
}

func (f *Metadata) dblog(err ...error) *logrus.Entry {
	_, file, line, _ := runtime.Caller(1)
	file = strings.TrimPrefix(file, os.Getenv("GOPATH")+"/src/github.com/ckeyer/tarofs/")
	fields := logrus.Fields{
		"module": "leveldb",
		"file":   fmt.Sprintf("%s:%v", file, line),
	}
	if len(err) > 0 && err[0] != nil {
		fields["error"] = err[0]
	}
	return logrus.WithFields(fields)
}
