package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path"

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
	h, _ := hashids.NewWithData(hd)
	userKey, _ := h.EncodeInt64([]int64{userId})
	hash := md5.Sum(photo)
	photoKey, err := h.EncodeInt64([]int64{
		int64(binary.BigEndian.Uint64(hash[:8]) & 0x7FFFFFFFFFFFFFFF),
		int64(binary.BigEndian.Uint64(hash[8:16]) & 0x7FFFFFFFFFFFFFFF),
	})
	if err != nil {
		return "", err
	}

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
