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
- clone the project `mkdir $GOPATH/src/github.com/aitour && cd $GOPATH/src/github.com/aitour && git clone https://github.com/aitour/scene.git`
- `cd scene`
- run `dep ensure`

## build and run
- `cd $GOPATH/src/github.com/aitour/scene/service && go build -o aiserver.exe && aiserver`
- in another shell, `cd $GOPATH/src/github.com/aitour/scene/web && go build -o aiweb.exe && aiweb`


## misc
- generate cert for grpc authenticate and ssl for http
```openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3650 -nodes```

- compile proto files
```protoc --go_out=plugins=grpc:. serverpb/rpc.proto```


