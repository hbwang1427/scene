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
- install git
- install dep `go get -u github.com/golang/dep/cmd/dep`
- clone the project `git clone https://github.com/aitour/scene.git`
- run `dep install`


## misc
- generate cert for grpc authenticate and ssl for http
```openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3650 -nodes```

- compile proto files
```protoc --go_out=plugins=grpc:. serverpb/rpc.proto```


