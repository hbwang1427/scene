package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	//for image decode
	_ "image/jpeg"
	_ "image/png"

	pb "github.com/aitour/scene/serverpb"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	conf   = flag.String("conf", "web.toml", "Specify a config file")
	config *Config
)

type Config struct {
	Http struct {
		Bind      string
		AssetsDir string
	}

	Grpc struct {
		Addr string
		Cert string
		Host string
	}
}

func parseConfig(conf string) (*Config, error) {
	var c Config
	content, err := ioutil.ReadFile(conf)
	if err != nil {
		return nil, err
	}
	if _, err = toml.Decode(string(content), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func getGrpcConn() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if len(config.Grpc.Cert) > 0 {
		creds, err := credentials.NewClientTLSFromFile(config.Grpc.Cert, config.Grpc.Host)
		if err != nil {
			return nil, fmt.Errorf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(config.Grpc.Addr, opts...)
	return conn, err
}

func createHttpServer() (*http.Server, error) {
	r := gin.Default()
	r.LoadHTMLGlob(config.Http.AssetsDir + "/templates/*")
	r.Static("/assets", config.Http.AssetsDir)
	r.GET("/demo", func(c *gin.Context) {
		c.HTML(http.StatusOK, "demo.html", gin.H{
			"title": "predict demo page",
		})
	})

	r.POST("/predict", func(c *gin.Context) {
		jpeg := c.PostForm("image")
		if strings.Index(jpeg, "data:image/jpeg;base64,") == 0 {
			jpeg = jpeg[23:]
		} else {
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
			AcquireVedio: false,
		})

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"text":  response.Text,
			"audio": response.AudioUrl,
		})
	})

	s := &http.Server{
		Addr:    config.Http.Bind,
		Handler: r,
	}
	return s, nil
}

func main() {
	flag.Parse()

	//parse config
	var err error
	if config, err = parseConfig(*conf); err != nil {
		log.Fatal(err)
	}
	log.Printf("%v", config)

	//create http server
	srv, err := createHttpServer()
	if err != nil {
		log.Fatal(err)
	}

	//startup http server
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
