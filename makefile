

target:
	go build
	./bemsupload.exe

export GOARCH=arm
export GOARM=5
export GOOS=linux
export CGO_ENABLED=1
export CC=arm-linux-gcc

arm:
	go build -ldflags "-s";
	wput -u -nc bemsupload ftp://172.18.5.34/bemsupload