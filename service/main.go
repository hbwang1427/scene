package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"syscall"

	pb "github.com/aitour/scene/serverpb"
	"golang.org/x/net/trace"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	_ "image/jpeg"
	_ "image/png"
)

var (
	conf                = flag.String("conf", "service.toml", "Specify the config file")
	config              *Config
	MissingWebHostError = errors.New("webhost config missing")
)

type Config struct {
	Grpc struct {
		Bind        string
		Cert        string
		Key         string
		TraceEnable bool
		TraceBind   string
	}

	Web struct {
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
	if len(c.Web.Host) == 0 {
		return nil, MissingWebHostError
	}
	c.Web.Host = strings.TrimRight(c.Web.Host, "/")

	return &c, nil
}

// server is used to implement serverpb.AuthServer
type authserver struct{}

// Authenticate implements serverpb.AuthServer.Authenticate
func (s *authserver) Authenticate(ctx context.Context, in *pb.AuthRequest) (*pb.AuthResponse, error) {
	var response = &pb.AuthResponse{}
	if (in.Name != "test" || in.Password != "123") && in.Token != "this is an valid token" {
		response.Msg = "invalid name or password or invalid token"
	} else {
		response.Token = "this is another valid token"
	}
	return response, nil
}

//predictserver is used to implement serverpb.PredictServer
type predictserver struct{}

//PredictPhoto implements serverpb.PredictServer
func (s *predictserver) PredictPhoto(ctx context.Context, in *pb.PhotoPredictRequest) (*pb.PhotoPredictResponse, error) {
	var response = &pb.PhotoPredictResponse{}

	// m, _, err := image.Decode(bytes.NewReader(in.Data))
	// if err != nil {
	// 	return nil, err
	// }
	// bounds := m.Bounds()
	// var histogram [16][4]int
	// for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
	// 	for x := bounds.Min.X; x < bounds.Max.X; x++ {
	// 		r, g, b, a := m.At(x, y).RGBA()
	// 		// A color's RGBA method returns values in the range [0, 65535].
	// 		// Shifting by 12 reduces this to the range [0, 15].
	// 		histogram[r>>12][0]++
	// 		histogram[g>>12][1]++
	// 		histogram[b>>12][2]++
	// 		histogram[a>>12][3]++
	// 	}
	// }

	// var b bytes.Buffer
	// fmt.Fprintf(&b, "%-14s %6s %6s %6s %6s\n", "bin", "red", "green", "blue", "alpha")
	// for i, x := range histogram {
	// 	fmt.Fprintf(&b, "0x%04x-0x%04x: %6d %6d %6d %6d\n", i<<12, (i+1)<<12-1, x[0], x[1], x[2], x[3])
	// }
	// response.Text = b.String()

	var audiourl string
	language := in.Language
	if language == "zh" {
		audiourl = fmt.Sprintf("%s/assets/audio/sample_0.4mb.mp3", config.Web.Host)
	} else if language == "de" {
		audiourl = fmt.Sprintf("%s/assets/audio/sample_0.4mb.mp3", config.Web.Host)
	} else {
		//other languages: de fr it ...
		audiourl = fmt.Sprintf("%s/assets/audio/sample_0.4mb.mp3", config.Web.Host)
	}

	response.Results = []*pb.PhotoPredictResponse_Result{&pb.PhotoPredictResponse_Result{
		Text:      "a person picture",
		ImageUrl:  fmt.Sprintf("%s/assets/imgs/c1.jpg", config.Web.Host),
		AudioUrl:  audiourl,
		AudioSize: 443926,
		AudioLen:  27,
	}, &pb.PhotoPredictResponse_Result{
		Text:     "dish",
		ImageUrl: fmt.Sprintf("%s/assets/imgs/c2.jpg", config.Web.Host),
	}, &pb.PhotoPredictResponse_Result{
		Text:     "building",
		ImageUrl: fmt.Sprintf("%s/assets/imgs/c3.jpg", config.Web.Host),
	}, &pb.PhotoPredictResponse_Result{
		Text:     "220px-Buckman_Tavern_Lexington_Massachusetts",
		ImageUrl: fmt.Sprintf("%s/assets/imgs/220px-Buckman_Tavern_Lexington_Massachusetts.jpg", config.Web.Host),
	}, &pb.PhotoPredictResponse_Result{
		Text:     "250px-Minute_Man_Statue_Lexington_Massachusetts",
		ImageUrl: fmt.Sprintf("%s/assets/imgs/250px-Minute_Man_Statue_Lexington_Massachusetts.jpg", config.Web.Host),
	}, &pb.PhotoPredictResponse_Result{
		Text:     "3178927_orig",
		ImageUrl: fmt.Sprintf("%s/assets/imgs/3178927_orig.jpg", config.Web.Host),
	}, &pb.PhotoPredictResponse_Result{
		Text:     "vt",
		ImageUrl: fmt.Sprintf("%s/assets/imgs/vt.jpg", config.Web.Host),
	}, &pb.PhotoPredictResponse_Result{
		Text:     "images3.jpeg",
		ImageUrl: fmt.Sprintf("%s/assets/imgs/images3.jpeg", config.Web.Host),
	}}

	if in.MaxLimits > 0 && len(response.Results) > int(in.MaxLimits) {
		response.Results = response.Results[:in.MaxLimits]
	}
	//time.Sleep(20 * time.Second)
	return response, nil
}

func createGrpcServer() (*grpc.Server, error) {
	//make credentials for grpc
	cert, err := tls.LoadX509KeyPair(config.Grpc.Cert, config.Grpc.Key)
	if err != nil {
		return nil, err
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   0,
	})

	// creds, err := credentials.NewServerTLSFromFile(config.Grpc.Cert, config.Grpc.Key)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to generate credentials %v", err)
	// }
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.MaxSendMsgSize(20 * 1024 * 1024), //max send message size set to 20MB
		grpc.MaxRecvMsgSize(20 * 1024 * 1024), //max recv message size set to 20MB
	}

	//create grpc server
	grpc.EnableTracing = config.Grpc.TraceEnable
	s := grpc.NewServer(opts...)

	//register serverpb.AuthServer
	pb.RegisterAuthServer(s, &authserver{})
	pb.RegisterPredictServer(s, &predictserver{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	return s, nil
}

func main() {
	flag.Parse()
	var err error
	if config, err = parseConfig(*conf); err != nil {
		log.Fatal(err)
	}

	//create grpc server
	s, err := createGrpcServer()
	if err != nil {
		log.Fatal(err)
	}

	//startup grpc server
	log.Printf("starting grpc server on %s", config.Grpc.Bind)
	lis, err := net.Listen("tcp", config.Grpc.Bind)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	//startup trace server if wanted
	//visit /debug/requests in browser to trace requests
	if config.Grpc.TraceEnable {
		trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
			return true, true
		}
		log.Printf("visit tracing at: %s", config.Grpc.TraceBind)
		go http.ListenAndServe(config.Grpc.TraceBind, nil)
	}

	//setup signal handlers to handle signals sent to stop this process
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	sig := <-quit
	log.Printf("signal %s received. shutdown server ...", sig.String())

	s.GracefulStop()

	log.Println("service stopped")
}
