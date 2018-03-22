export GOARCH=arm
export GOARM=5
export GOOS=linux

target:
	go build -ldflags "-s"
	cp bemsupload /c/ftp