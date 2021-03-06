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

serverpb/rpc.pb.go: serverpb/rpc.proto
	@protoc -I serverpb/ \
			--go_out=plugins=grpc:. \
			serverpb/rpc.proto

protogen: serverpb/rpc.pb.go


build-server: protogen
	@echo "building aiserver"
	@go build -o $(SERVER_OUT) $(SERVER_PKG)

build-web: protogen
	@echo "building aiweb"
	@go build -o $(WEB_OUT) $(WEB_PKG)

cpresource:
	@/bin/cp -r certs $(DEPLOY_DIR)/
	@/bin/cp -r web/assets $(DEPLOY_DIR)/aiweb
	@/bin/cp supervisord.ini $(DEPLOY_DIR)/

clean:
	@rm $(SERVER_OUT) $(WEB_OUT)

all: build-server build-web cpresource 

