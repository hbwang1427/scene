package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aitour/scene/model"
	"github.com/aitour/scene/web/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

func UploadAnonomousePhoto(c *gin.Context) {
	file, _ := c.FormFile("file")
	fileKey := strings.Split(c.PostForm("filekey"), ",")
	if len(fileKey) != 2 {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	id1, _ := strconv.ParseUint(fileKey[0], 10, 32)
	id2, _ := strconv.ParseUint(fileKey[1], 10, 32)
	sid := int64(id1)<<32 + int64(id2)
	fileURL := fmt.Sprintf("%d_%d.bmp", time.Now().UnixNano(), sid)
	dst := path.Join(config.GetConfig().Http.UploadDir, fileURL)

	//save upload file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	defer src.Close()

	//calc the md5
	md5h := md5.New()
	io.Copy(md5h, src)
	fileHash := fmt.Sprintf("%x", md5h.Sum([]byte("")))

	//c.SaveUploadedFile(file)
	uploadFrom := c.Request.RemoteAddr
	if i := strings.Index(uploadFrom, ":"); i > 0 {
		uploadFrom = uploadFrom[:i]
	}
	err = model.InsertPorcelainPhoto(sid, fileURL, -1, uploadFrom, fileHash)
	if err != nil {
		log.Printf("upload anonomouse photo error:%v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	//if md5 was not violate the filehash unique constraint then we save it to disk
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	defer out.Close()

	io.Copy(out, src)
}

func SetAnonomouseUploadedPhotoClass(c *gin.Context) {
	fileKey := strings.Split(c.PostForm("filekey"), ",")
	if len(fileKey) != 2 {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	id1, _ := strconv.ParseUint(fileKey[0], 10, 32)
	id2, _ := strconv.ParseUint(fileKey[1], 10, 32)
	sid := int64(id1)<<32 + int64(id2)
	class, err := strconv.Atoi(c.PostForm("class"))
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	err = model.UpdatePorcelainClass(sid, uint16(class))
	if err != nil {
		log.Printf("update photo class error:%v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
}
