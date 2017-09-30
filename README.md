# AiTour scene project

## structure
- service 
sample service based on grpc

- client
cli client demo to call service

- web
http server demo. 


## usage
- download and install [golang](https://golang.org/dl/)
- config GOROOT and GOPATH environment variable, append $GOPATH/bin to PATH environment variable
- install git
- install dep `go get -u github.com/golang/dep/cmd/dep`
- clone the project `cd $GOPATH/src/github.com/ && git clone https://github.com/aitour/scene.git`
- `cd aitour\scene`
- run `dep ensure`


## misc
- generate cert for grpc authenticate and ssl for http
```openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3650 -nodes```

- compile proto files
```protoc --go_out=plugins=grpc:. serverpb/rpc.proto```


