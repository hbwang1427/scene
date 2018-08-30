package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/aitour/scene/model"
	pb "github.com/aitour/scene/serverpb"
	"github.com/aitour/scene/web/config"
)

var (
	cfg *config.Config

	supportedImageTypes = []struct {
		header string
		ptype  pb.PhotoPredictRequest_PhotoType
	}{
		{"data:image/jpg;base64,", pb.PhotoPredictRequest_JPG},
		{"data:image/jpeg;base64,", pb.PhotoPredictRequest_JPG},
		{"data:image/png;base64,", pb.PhotoPredictRequest_PNG},
	}

	imgTypes = map[string]pb.PhotoPredictRequest_PhotoType{
		"jpg":  pb.PhotoPredictRequest_JPG,
		"jpeg": pb.PhotoPredictRequest_JPG,
		"png":  pb.PhotoPredictRequest_PNG,
	}

	testImages = make(map[string][]string)
)

func getGrpcConn() (*grpc.ClientConn, error) {
	if cfg == nil {
		cfg = config.GetConfig()
	}
	var opts []grpc.DialOption
	if len(cfg.Grpc.Cert) > 0 {
		creds, err := credentials.NewClientTLSFromFile(cfg.Grpc.Cert, cfg.Grpc.Host)
		if err != nil {
			return nil, fmt.Errorf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(cfg.Grpc.Addr, opts...)
	return conn, err
}

//curl -X POST http://localhost:8081/predict -F "image=@C:\Users\kingwang\Desktop\PNG_transparency_demonstration_1.png" -F "imgtype=jpeg"
func Predict(c *gin.Context) {
	var imgData []byte
	var err error
	var lat = LatError
	var lng = LngError

	//extract limits. default limits set to 10
	limits, _ := strconv.Atoi(c.DefaultPostForm("limits", "10"))
	if limits > 10 {
		limits = 10
	}

	lat, _ = strconv.ParseFloat(c.PostForm("lat"), 64)
	lng, _ = strconv.ParseFloat(c.PostForm("lng"), 64)
	//log.Printf("lat:%f, lng:%f", lat, lng)
	language := c.PostForm("language")
	log.Printf("language: %s", language)
	site := c.PostForm("site")
	log.Printf("site: %s", site)

	//extract image type
	var imgType pb.PhotoPredictRequest_PhotoType
	if v := c.PostForm("imgtype"); len(v) > 0 {
		//it's a multipart encoded request
		var ok bool
		imgType, ok = imgTypes[v]
		if !ok {
			//image type not acceptable
			c.JSON(http.StatusOK, gin.H{
				"error": "image type not support",
			})
			return
		}

		//extract image data
		image, err := c.FormFile("image")
		if err != nil {
			//image type not acceptable
			c.JSON(http.StatusOK, gin.H{
				"error": "invalid image",
			})
			return
		}

		f, _ := image.Open()
		imgData, _ = ioutil.ReadAll(f)
		f.Close()
	} else {
		b64EncodedImage := c.PostForm("image")
		//data:image/png;base64,
		if i := strings.Index(b64EncodedImage, ";base64,"); i > 0 {
			for _, t := range supportedImageTypes {
				if i := strings.Index(b64EncodedImage, t.header); i >= 0 {
					imgType = t.ptype
					b64Decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64EncodedImage[len(t.header):]))
					if imgData, err = ioutil.ReadAll(b64Decoder); err != nil {
						c.JSON(http.StatusOK, gin.H{
							"error": "invalid image",
						})
						return
					}
				}
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error": "invalid image",
			})
			return
		}
	}

	//save user image
	if v, ok := c.Get(gin.AuthUserKey); ok {
		uid, _ := strconv.ParseInt(v.(string), 10, 64)
		photoStore := NewDiskPhotoStore(config.GetConfig().Http.UploadDir)
		url, err := photoStore.Store(uid, imgData)
		if err == nil {
			photo := &model.UserAlbumPhoto{UserId: uid, Url: url}
			err := model.AddPhoto(photo)
			if err != nil {
				log.Printf("add photo error:%v", err)
			}
		} else {
			log.WithError(err).Error("store image error")
		}
	}

	// get a connection to the server.
	conn, err := getGrpcConn()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "temporarily out of service",
		})
	}
	defer conn.Close()

	//create client stub
	client := pb.NewPredictClient(conn)
	response, err := client.PredictPhoto(context.Background(), &pb.PhotoPredictRequest{
		Type:         imgType,
		Data:         imgData,
		Language:     language,
		Site:         site,
		Geo:          &pb.GeoPosition{Latitude: lat, Longitude: lng},
		AcquireText:  true,
		AcquireAudio: true,
		AcquireVideo: false,
		MaxLimits:    int32(limits),
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": response.Results,
	})
}

func resizeImage(fname string) (string, error) {
	file, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fnameLower := strings.ToLower(fname)
	var img image.Image
	if strings.HasSuffix(fnameLower, "jpg") || strings.HasSuffix(fnameLower, "jpeg") {
		if img, err = jpeg.Decode(file); err != nil {
			return "", err
		}
	} else {
		if img, err = png.Decode(file); err != nil {
			return "", err
		}
	}

	var resizedImage image.Image
	if img.Bounds().Dx() > img.Bounds().Dy() {
		resizedImage = resize.Resize(600, 0, img, resize.Lanczos3)
	} else {
		resizedImage = resize.Resize(0, 600, img, resize.Lanczos3)
	}

	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, resizedImage, nil); err != nil {
		return "", err
	}

	contents, err := ioutil.ReadAll(out)
	if err != nil {
		return "", err
	}

	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(contents), nil
}

func FetchTestImages(c *gin.Context) {
	site := c.Param("site")
	if len(site) == 0 {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	if v, ok := testImages[site]; ok {
		c.JSON(http.StatusOK, gin.H{
			"images": v,
		})
		return
	}

	testImages[site] = make([]string, 0)
	fis, err := ioutil.ReadDir("assets/demo/" + site)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
	}
	for _, fi := range fis {
		if !fi.IsDir() {
			base64EncodedImage, err := resizeImage(path.Join("assets/demo", site, fi.Name()))
			if err == nil {
				testImages[site] = append(testImages[site], base64EncodedImage)
			} else {
				log.WithError(err).Error("load image error")
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"images": testImages[site],
	})
	return
}
