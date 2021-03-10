package weedfs

import (
	"github.com/chrislusf/seaweedfs/weed/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type WeedFS struct {
	log       *logrus.Entry
	weedStore *storage.Store
}

func NewWeedFS() (*WeedFS, error) {
	s := storage.NewStore(
		grpc.WithInsecure(), 0, "", "",
		[]string{},
		[]int{},
		[]float32{},
		"/tmp/weedfs/idxdir",
		storage.NeedleMapInMemory,
	)
	wfs := &WeedFS{
		weedStore: s,
	}
	return wfs, nil
}

func (wfs *WeedFS) run() {

}
