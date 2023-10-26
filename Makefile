SHELL:=/bin/bash
.ONESHELL:

# apply code formatting
clean:
	go mod tidy
	gofmt -l -w .

SRC:=main.go
BIN:=fastqSplit
FASTQ:=data/test1.fastq
FASTQGZ:=data/test1.fastq.gz
test-run:
	cat $(FASTQ) | go run $(SRC)
	gunzip -c $(FASTQGZ) | go run $(SRC)
	go run $(SRC) <(gunzip -c $(FASTQGZ))
	go run $(SRC) $(FASTQ)
	go run $(SRC) $(FASTQGZ)

# NOTE: you can just ignore this error message;
# fatal: No names found, cannot describe anything.
GIT_TAG:=$(shell git describe --tags)

build:
	go build -ldflags="-X 'main.Version=$(GIT_TAG)'" -o ./$(BIN) ./$(SRC)
.PHONY:build

build-all:
	mkdir -p build ; \
	for os in darwin linux windows; do \
	for arch in amd64 arm64; do \
	output="build/$(BIN)-v$(GIT_TAG)-$$os-$$arch" ; \
	if [ "$${os}" == "windows" ]; then output="$${output}.exe"; fi ; \
	echo "building: $$output" ; \
	GOOS=$$os GOARCH=$$arch go build -ldflags="-X 'main.Version=$(GIT_TAG)'" -o "$${output}" $(SRC) ; \
	done ; \
	done


build-test-run: build
	cat $(FASTQ) | ./$(BIN)
	gunzip -c $(FASTQGZ) | ./$(BIN)
	./$(BIN) <(gunzip -c $(FASTQGZ))
	./$(BIN) $(FASTQ)
	./$(BIN) $(FASTQGZ)

# docker build -t stevekm/fastq-split:latest .
DOCKER_TAG:=stevekm/fastq-split:$(GIT_TAG)
docker-build:
	docker build --build-arg "Version=$(GIT_TAG)" -t $(DOCKER_TAG) .

# docker push stevekm/fastq-split:latest
docker-push:
	docker push $(DOCKER_TAG)

docker-test-run:
	docker run --platform linux/amd64 --rm -ti -v ${PWD}:${PWD} --workdir ${PWD} $(DOCKER_TAG) $(BIN) $(FASTQ)