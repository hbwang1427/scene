package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
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

func init() {
	if cfg == nil {
		cfg = config.GetConfig()
	}
}

func getGrpcConn() (*grpc.ClientConn, error) {
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
	var uid int64 = -1
	if v, ok := c.Get(gin.AuthUserKey); ok {
		uid, _ = strconv.ParseInt(v.(string), 10, 64)
	}

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

type ArtScore struct {
	ArtID int
	Score float64
}

//Predict2 feature match
//client extracts feature in client side and pass it to server to match the result
func Predict2(c *gin.Context) {
	var feature []float64
	for {
		var v float32
		err := binary.Read(c.Request.Body, binary.LittleEndian, &v)
		if err != nil {
			break
		}
		
		feature = append(feature,float64(v))
	}
	log.Printf("%v %v %v", feature[2], feature[3], feature[len(feature)-1])
	//log.Printf("%v", feature)
	refers, err := model.GetArtReferences()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err,
		})
		return
	}
	if len(refers) == 0 || len(refers[0].MobileNetFeature) != len(feature) {
		c.JSON(http.StatusOK, gin.H{
			"err": "no reference feature or feature length not match",
		})
		log.Printf("refers len:%d, image_feature len:%d, feature len:%d", 
			len(refers), len(refers[0].MobileNetFeature), len(feature))
		return
	}

	k := 5
	if arg := c.Query("k"); len(arg) > 0 {
		if v , err := strconv.Atoi(arg); err == nil {
			k = v
		}
	}
	topK := make([]ArtScore, k)
	featureNorm := model.Norm(feature)
    //var scores []float64
	for _, ref := range refers {
		var score float64
		for i := 0; i < len(ref.MobileNetFeature); i++ {
			score += ref.MobileNetFeature[i] * feature[i]
		}
		//score = score  / (ref.MobileNetFeatureNorm * featureNorm)
        score = score / featureNorm
        //scores = append(scores, score)

		for j := k-1; j >= 0; j-- {
            if j == 0  && score > topK[0].Score {
                topK =  append(append([]ArtScore{}, ArtScore{ref.ArtID, score}), topK[0:k-1]...)
            } else if j>0 && score > topK[j].Score && score <= topK[j-1].Score {
                topK = append(append(append([]ArtScore{}, topK[0:j]...), ArtScore{ref.ArtID, score}), topK[j:k-1]...)
				break
			}
		}
	}
    //sort.Slice(scores, func(i, j int) bool {
    //    return scores[i] > scores[j]
    //}) 
    //ioutil.WriteFile("/usr/local/aitour/aiweb/scores.txt", []byte(fmt.Sprintf("%v", scores)), 0644)

	c.JSON(http.StatusOK, gin.H{
		"results": topK,
	})
}


func GetArtById(c *gin.Context) {
	artid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "invalid request param",
		})
		return
	}
	language_id, err := strconv.Atoi(c.Query("language"))
	if err != nil {
		language_id = 2
	}
        
        assets_path := "/assets/MET/"
        art, err := model.GetArtById(artid, language_id)

	if err != nil {
		log.Printf("err:%v", err)
		c.JSON(http.StatusOK, gin.H{
			"error": "art not exist",
		})
		return
	}

	for i := 0; i < len(art.Images); i++ {
		art.Images[i] = assets_path + "Images/" + art.Images[i] + ".jpg"
	  }
	  for i := 0; i < len(art.Audios); i++ {
		art.Audios[i] = assets_path + "Audio/" + art.Audios[i]
	  }
	  
	c.JSON(http.StatusOK, gin.H{
		"results": art,
	})
}
