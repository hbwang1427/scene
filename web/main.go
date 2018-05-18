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

	"github.com/aitour/scene/web/config"
	"github.com/aitour/scene/web/handler"
	"github.com/gin-contrib/cors"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

var (
	conf = flag.String("conf", "web.toml", "Specify a config file")
	cfg  *config.Config
)

func createHttpServer() (*http.Server, error) {
	log.SetOutput(gin.DefaultWriter)

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
	r.Static("/photo", cfg.Http.UploadDir)

	r.GET("/user/register", func(c *gin.Context) {
		id := captcha.New()
		c.HTML(http.StatusOK, "register.html", gin.H{
			"cv": id,
		})
	})
	r.POST("/user/register", handler.CreateUser)
	r.GET("/user/activate", handler.ActivateUser)
	r.GET("/user/signin", handler.UserLogin)
	r.POST("/user/signin", handler.AuthUser)
	r.GET("/user/logout", handler.Logout)
	authorized := r.Group("/user", handler.AuthChecker())
	authorized.GET("/profile", func(c *gin.Context) {
		fmt.Fprintf(c.Writer, "when you see this page, you have passed the auth check!")
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "pangolins ai",
		})
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
	r.GET("/vcode/:img", gin.WrapH(captcha.Server(200, 60)))
	r.GET("/vcode", handler.NewCaptacha)

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
