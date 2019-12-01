package levelfs

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ckeyer/tarofs/pkgs/storage"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

var _ storage.MetadataStorager = (*leveldbStorage)(nil)
var _ storage.DataStorager = (*leveldbStorage)(nil)

type leveldbStorage struct {
	mlog *logrus.Logger
	db   *leveldb.DB
}

func NewLevelStorage(leveldir string) (*leveldbStorage, error) {
	m := &leveldbStorage{mlog: logrus.New()}
	m.mlog.SetLevel(logrus.WarnLevel)

	db, err := leveldb.OpenFile(leveldir, nil)
	if err != nil {
		return nil, err
	}

	m.db = db

	return m, nil
}

func (l *leveldbStorage) Bytes(key string) ([]byte, error) {
	bs, err := l.db.Get([]byte(key), nil)
	if err != nil {
		if leveldb.ErrNotFound == err {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return bs, nil
}

func (f *leveldbStorage) PutBytes(key string, val []byte) error {
	return f.db.Put([]byte(key), val, nil)
}

func (f *leveldbStorage) Get(key string, ret interface{}) error {
	val, err := f.Bytes(key)
	if err != nil {
		f.dblog(err).
			WithField("key", key).
			Infof("get failed.")
		return err
	}
	f.dblog().
		WithField("key", key).
		Debugf("get %s", val)

	if ret == nil {
		return nil
	}

	if err := jsonDecode(val, ret); err != nil {
		return err
	}
	return nil
}

func (f *leveldbStorage) Put(key string, val interface{}) error {
	data := jsonEncode(val)
	err := f.PutBytes(key, data)
	if err != nil {
		f.dblog(err).
			WithField("data", string(data)).
			WithField("key", key).
			Debugf("put failed.")
		return err
	}
	f.dblog().
		WithField("data", string(data)).
		WithField("key", key).
		Debugf("put successful.")
	return nil
}

// delete
func (f *leveldbStorage) Delete(key string) error {
	err := f.db.Delete([]byte(key), nil)
	if err != leveldb.ErrNotFound {
		f.dblog(err).
			WithField("key", key).
			Debugf("delete not found.")
		return nil
	} else if err != nil {
		f.dblog(err).
			WithField("key", key).
			Debugf("delete failed.")
		return err
	}
	f.dblog().
		WithField("key", key).
		Debugf("delete successful.")
	return nil
}

func (f *leveldbStorage) Close() error {
	return f.db.Close()
}

func jsonEncode(val interface{}) []byte {
	bs, _ := json.Marshal(val)
	return bs
}

func jsonDecode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (f *leveldbStorage) dblog(err ...error) *logrus.Entry {
	_, file, line, _ := runtime.Caller(1)
	file = strings.TrimPrefix(file, os.Getenv("GOPATH")+"/src/github.com/ckeyer/tarofs/")
	fields := logrus.Fields{
		"module": "leveldb",
		"file":   fmt.Sprintf("%s:%v", file, line),
	}
	if len(err) > 0 && err[0] != nil {
		fields["error"] = err[0]
	}

	return f.mlog.WithFields(fields)
}
