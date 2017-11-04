package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/aitour/scene/serverpb"
	"github.com/aitour/scene/web/config"
)

var (
	cfg *config.Config
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

func Predict(c *gin.Context) {
	jpeg := c.PostForm("image")
	if i := strings.Index(jpeg, ";base64,"); i > 0 {
		jpeg = jpeg[i+len(";base64,"):]
	} else {
		log.Printf("invalid image:%s", jpeg)
		c.JSON(http.StatusOK, gin.H{
			"error": "invalid image",
		})
		return
	}

	// get a connection to the server.
	conn, err := getGrpcConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	imgdata, err := ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(jpeg)))
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusOK, gin.H{
			"error": "invalid image",
		})
		return
	}

	//create client stub
	client := pb.NewPredictClient(conn)
	response, err := client.PredictPhoto(context.Background(), &pb.PhotoPredictRequest{
		Type:         pb.PhotoPredictRequest_PNG,
		Data:         imgdata,
		Geo:          &pb.GeoPosition{},
		AcquireText:  true,
		AcquireAudio: true,
		AcquireVideo: false,
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
