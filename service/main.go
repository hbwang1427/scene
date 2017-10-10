package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"

	"syscall"

	pb "github.com/aitour/scene/serverpb"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	_ "image/jpeg"
	_ "image/png"
)

var (
	conf   = flag.String("conf", "service.toml", "Specify the config file")
	config *Config
)

type Config struct {
	Grpc struct {
		Bind string
		Cert string
		Key  string
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

	m, _, err := image.Decode(bytes.NewReader(in.Data))
	if err != nil {
		return nil, err
	}
	bounds := m.Bounds()
	var histogram [16][4]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := m.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].
			histogram[r>>12][0]++
			histogram[g>>12][1]++
			histogram[b>>12][2]++
			histogram[a>>12][3]++
		}
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "%-14s %6s %6s %6s %6s\n", "bin", "red", "green", "blue", "alpha")
	for i, x := range histogram {
		fmt.Fprintf(&b, "0x%04x-0x%04x: %6d %6d %6d %6d\n", i<<12, (i+1)<<12-1, x[0], x[1], x[2], x[3])
	}
	response.Text = b.String()
	response.AudioUrl = "http://disney616.com:8081/assets/audio/sample_0.4mb.mp3"

	return response, nil
}

func createGrpcServer() (*grpc.Server, error) {
	//make credentials for grpc
	creds, err := credentials.NewServerTLSFromFile(config.Grpc.Cert, config.Grpc.Key)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate credentials %v", err)
	}
	opts := []grpc.ServerOption{grpc.Creds(creds)}

	//create grpc server
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

	//setup signal handlers to handle signals sent to stop this process
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	sig := <-quit
	log.Printf("signal %s received. shutdown server ...", sig.String())

	s.GracefulStop()

	log.Println("service stopped")
}
