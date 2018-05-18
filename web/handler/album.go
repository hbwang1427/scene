package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	hashids "github.com/speps/go-hashids"
)

// image formats and magic numbers
var magicTable = map[string]string{
	"\xff\xd8\xff":      "jpeg",
	"\x89PNG\r\n\x1a\n": "png",
	// "GIF87a":            "gif",
	// "GIF89a":            "gif",
}

type PhotoStore interface {
	Store(userId int64, photo []byte) (url string, err error)
}

type DiskPhotoStore struct {
	storeRoot string
}

func (store *DiskPhotoStore) Store(userId int64, photo []byte) (url string, err error) {
	//check photo type
	var photoSuffix string
	for header, suffix := range magicTable {
		if bytes.HasPrefix(photo, []byte(header)) {
			photoSuffix = suffix
			break
		}
	}
	if len(photoSuffix) == 0 {
		return "", fmt.Errorf("Invalid photo type")
	}

	hd := hashids.NewData()
	// hd.Salt = "you guess"
	// hd.MinLength = 30
	h, _ := hashids.NewWithData(hd)
	userKey, _ := h.EncodeInt64([]int64{userId})
	photoKey, _ := h.EncodeInt64([]int64{time.Now().UnixNano()})

	url = fmt.Sprintf("%s/%s.%s", userKey, photoKey, photoSuffix)
	err = nil

	//write to disk

	os.MkdirAll(path.Join(store.storeRoot, userKey), 0777)
	err = ioutil.WriteFile(path.Join(store.storeRoot, url), photo, 0600)
	url = "/photo/" + url
	return
}

func NewDiskPhotoStore(storeRoot string) *DiskPhotoStore {
	return &DiskPhotoStore{storeRoot}
}
