package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "github.com/aitour/scene/serverpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	serverAddr = flag.String("serveraddr", "localhost:50051", "Grpc server address")
	caCert     = flag.String("cacert", "../certs/cert.pem", "The certificate of Grpc server")
	hostName   = flag.String("host", "disney616.com", "The host name to match the certificate")
)

func getGrpcConn() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if len(*caCert) > 0 {
		creds, err := credentials.NewClientTLSFromFile(*caCert, *hostName)
		if err != nil {
			return nil, fmt.Errorf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*serverAddr, opts...)
	return conn, err
}

func main() {
	flag.Parse()

	// get a connection to the server.
	conn, err := getGrpcConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewAuthClient(conn)

	// bad auth
	r, err := c.Authenticate(context.Background(), &pb.AuthRequest{Name: "wangxr", Password: "12345"})
	if err != nil {
		log.Fatalf("could not auth: %v", err)
	}
	log.Printf("auth: token=%s, msg=%s", r.Token, r.Msg)

	//valid auth by username and password
	r, err = c.Authenticate(context.Background(), &pb.AuthRequest{Name: "wangxr", Password: "123"})
	if err != nil {
		log.Fatalf("could not auth: %v", err)
	}
	log.Printf("auth: token=%s, msg=%s", r.Token, r.Msg)

	//valid auth by token
	r, err = c.Authenticate(context.Background(), &pb.AuthRequest{Token: "this is an valid token"})
	if err != nil {
		log.Fatalf("could not auth: %v", err)
	}
	log.Printf("auth: token=%s, msg=%s", r.Token, r.Msg)
}
