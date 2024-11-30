all: pre
	$(shell pwd)/build/build -o $(shell pwd)/build -i $(shell pwd)/cmd/anchorage -f anchorage

pre:
	mkdir -p ./build;
	go build -o $(shell pwd)/build/build -ldflags "-s -w" $(shell pwd)/script/build.go;

zip:
	mkdir -p ./_pack;
	zip -r _pack/anchorage.zip  build -x build/build
