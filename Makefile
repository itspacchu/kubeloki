all: build

build:
	go build -o ./bin/kubeloki .

clean:
	rm -rf ./bin/*