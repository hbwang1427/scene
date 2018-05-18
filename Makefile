PKG:= "github.com/aitour/scene"
SERVER_PKG:= "${PKG}/service"
WEB_PKG:= "${PKG}/web"
DEPLOY_DIR:="/usr/local/aitour"
SERVER_OUT:= "${DEPLOY_DIR}/aiserver/aiserver"
WEB_OUT:= "${DEPLOY_DIR}/aiweb/aiweb"
GO:= $(shell command -v go 2> /dev/null)
DEP:= $(shell command -v dep 2> /dev/null)

.PHONY: check dep cpresource protogen server web all clean

check:
ifndef GO
	$(error "please install golang first")
endif
ifeq ($(GOPATH),)
	$(error "please set GOPATH env variable first")
endif
	@if [ ! -d "${GOPATH}/src/github.com/golang/dep/cmd/dep" ]; then \
		echo "fetching go dep..."; \
		go get -u github.com/golang/dep/cmd/dep; \
	fi

dep: check
	$(GOPATH)/bin/dep ensure


serverpb/rpc.pb.go: serverpb/rpc.proto
	@protoc -I serverpb/ \
			--go_out=plugins=grpc:./serverpb \
			serverpb/rpc.proto

protogen: serverpb/rpc.pb.go


server: protogen
	@echo "building aiserver"
	@go build -o $(SERVER_OUT) $(SERVER_PKG)

web: protogen
	@echo "building aiweb"
	@go build -o $(WEB_OUT) $(WEB_PKG)

cpresource:
	@/bin/cp -r certs $(DEPLOY_DIR)/
	@/bin/cp service/service.toml $(DEPLOY_DIR)/aiserver/
	@/bin/cp web/web.toml  $(DEPLOY_DIR)/aiweb/
	@/bin/cp -r web/assets $(DEPLOY_DIR)/aiweb
	@/bin/cp supervisord.ini $(DEPLOY_DIR)/

clean:
	@rm $(SERVER_OUT) $(WEB_OUT)

all: dep server web

