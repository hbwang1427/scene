package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

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

	//extract limits. default limits set to 10
	limits, _ := strconv.Atoi(c.DefaultPostForm("limits", "10"))
	if limits > 10 {
		limits = 10
	}

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

	photoStore := NewDiskPhotoStore(config.GetConfig().Http.UploadDir)
	url, err := photoStore.Store(1781143536087860227, imgData)
	log.Printf("%v %v", url, err)

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
		Geo:          &pb.GeoPosition{},
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
