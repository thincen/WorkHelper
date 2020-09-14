# 名称
BINARY_NAME=WorkHelper
# 版本号
VERSION=0.1-beta.1
# 编译时间
DATE=`date +%F`
# 减小体积
LDFLAGS=-ldflags "-s -w -X 'main.VERSION=${VERSION}' -X 'main.NAME=${BINARY_NAME}' -H=windowsgui"

all: fmt build

fmt:
	go fmt ./...

build: win64 win32
# linux amd64
#linux-amd64:
#	@echo "building linux-amd64"
#	go build ${LDFLAGS} -o ${BINARY_NAME}_Linux_amd64 -v
#	upx -9 ${BINARY_NAME}_Linux_amd64

# win64版
win64:
	@echo "编译win64版本"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ./bin/${BINARY_NAME}_win_amd64.exe ./gui
# upx压缩 -9 压缩最佳体积
	upx -9 ./bin/$(BINARY_NAME)_win_amd64.exe

win32:
	@echo "building win32"
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build ${LDFLAGS} -o ./bin/${BINARY_NAME}_win_386.exe ./gui
	upx -9 ./bin/$(BINARY_NAME)_win_386.exe

zip:
	@echo "打包 ..."
	zip -r -9 ./release/$(BINARY_NAME)_win_386.zip ./bin/logo48.ico ./bin/template ./bin/$(BINARY_NAME)_win_386.exe
	zip -r -9 ./release/$(BINARY_NAME)_win_amd64.zip ./bin/logo48.ico ./bin/template ./bin/$(BINARY_NAME)_win_amd64.exe
