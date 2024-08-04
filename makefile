all: b

b:
	go mod tidy
	go build -o ./build -tags fastjson

r:
	go mod tidy
	go build -o ./build -ldflags "-s -w"

r-fast:
	go mod tidy
	go build -o ./build -ldflags "-s -w" -tags fastjson

run: b
	./build/lethalmodder $(var)

clean: 
	rm -rf ./build/* || true
