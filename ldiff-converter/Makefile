APP_NAME = ldiff-converter.exe

all: build

build:
	set GOOS=windows&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -buildvcs=false -trimpath -ldflags="-s -w" -o $(APP_NAME) ./

run: build
	./$(APP_NAME)

clean:
	rm -f $(APP_NAME)
