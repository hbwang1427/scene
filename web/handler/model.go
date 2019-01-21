package handler

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

//import log "github.com/sirupsen/logrus"

var modelInfos []ModelInfo

type ModelInfo struct {
	Name            string
	Md5Hash         string
	FileSizeInBytes int64
	DownloadPath    string
}

func reloadModelInfos() ([]ModelInfo, error) {
	var models  []ModelInfo
	dir := "./assets/tflite_models"
	if fis, err := ioutil.ReadDir(dir); err == nil {
		for _, fi := range fis {
			if !fi.IsDir() {
				if fr, err := os.Open(path.Join(dir, fi.Name())); err == nil {
					defer fr.Close()
					h := md5.New()
					if _, err := io.Copy(h, fr); err != nil {
						return nil, err
					}
					models = append(models, ModelInfo{
						Name:            fi.Name(),
						Md5Hash:         fmt.Sprintf("%x", h.Sum(nil)),
						FileSizeInBytes: fi.Size(),
						DownloadPath:    "/assets/tflite_models/" + fi.Name(),
					})
				}
			}
		}
	} else {
		return nil, err
	}

	return models, nil
}

func RefreshModelInfo(c *gin.Context) {
	if c.Query("auth") != "123456" {
		c.String(http.StatusOK, "you don't have permission to do this")
		return
	}

	if models, err := reloadModelInfos(); err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		modelInfos = models
		c.JSON(http.StatusOK, gin.H{
			"models": models,
		})
	}
}

func GetModelInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"models": modelInfos,
	})
}
