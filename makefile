all: b

b: generate
	go mod tidy
	go build -o ./build

r: generate
	@if [ -z "$(system)" ]; then \
		echo "No system specified"; \
		exit 1; \
	fi; \
	out=build/LethalModder-$(system); \
	[ "$(system)" = "windows" ] && out=$${out}.exe; \
	go build -o $${out} -ldflags="-s -w"

r-fast: generate
	go mod tidy
	go build -o ./build -ldflags "-s -w" -tags fastjson

run: b
	./build/lethalmodder $(var)

clean: 
	find ./build/ -maxdepth 1 ! -name "README.md" -type f -exec rm -rf {} +

generate:
	go run github.com/tc-hib/go-winres@latest make --product-version=git-tag --file-version=git-tag
