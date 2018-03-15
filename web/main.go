package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	//for image decode
	_ "image/jpeg"
	_ "image/png"

	"github.com/aitour/scene/auth"
	"github.com/aitour/scene/web/config"
	"github.com/aitour/scene/web/handler"
	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

var (
	conf          = flag.String("conf", "web.toml", "Specify a config file")
	cfg           *config.Config
	tokenProvider auth.TokenProvider
)

func createHttpServer() (*http.Server, error) {
	log.SetOutput(gin.DefaultWriter)
	var err error
	tokenProvider, err = auth.CreateTokenProvider("simple", map[string]interface{}{
		"tokenTTL": 10 * time.Second,
		"tokenLen": 16,
	})
	if err != nil {
		log.Fatalf("create TokenProvider error:%v", err)
	}

	r := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	//r.MaxMultipartMemory = 8 << 20 // 8 MiB

	//cross domain request config.
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		//AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return origin == "*"
		// },
		MaxAge: 12 * time.Hour,
	}))

	r.LoadHTMLGlob(cfg.Http.AssetsDir + "/templates/*")
	r.Static("/assets", cfg.Http.AssetsDir)

	authorized := r.Group("/user", handler.AuthChecker(tokenProvider))
	authorized.GET("/profile", func(c *gin.Context) {
		fmt.Fprintf(c.Writer, "when you see this page, you have passed the auth check!")
	})

	r.GET("/demo", func(c *gin.Context) {
		c.HTML(http.StatusOK, "demo.html", gin.H{
			"title": "predict demo page",
		})
	})

	r.POST("/predict", handler.Predict)
	r.GET("/weather/current", handler.GetCurrentWeather)
	r.GET("/weather/forecast", handler.GetWeatherForeCast)
	r.GET("/geocode", handler.GeoCodeHandler)
	r.GET("/nearby/city", handler.FindNearbyCityHandler)
	r.GET("/nearby/museum", handler.SearchNearbyMuseumsByGoogleMap)
	r.GET("/place/photo", handler.GetPlacePhoto)
	r.GET("/place/detail", handler.GetPlaceDetail)

	s := &http.Server{
		Addr:    cfg.Http.Bind,
		Handler: r,
	}
	return s, nil
}

func main() {
	flag.Parse()

	//parse config
	var err error
	config.SetConfigPath(*conf)
	cfg = config.GetConfig()

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
