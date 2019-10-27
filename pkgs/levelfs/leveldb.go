package levelfs

import "github.com/syndtr/goleveldb/leveldb"

func (f *Metadata) getValue(key string) ([]byte, error) {
	return f.db.Get([]byte(key), nil)
}

func (f *Metadata) get(key string, ret interface{}) error {
	val, err := f.getValue(key)
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

func (f *Metadata) putValue(key string, val []byte) error {
	return f.db.Put([]byte(key), val, nil)
}

func (f *Metadata) put(key string, val interface{}) error {
	data := jsonEncode(val)
	err := f.putValue(key, data)
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
func (f *Metadata) delete(key string) error {
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
