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

build:
	go build -o ./$(BIN) ./$(SRC)
.PHONY:build

# fatal: No names found, cannot describe anything.
GIT_TAG:=$(shell git describe --tags)
build-all:
	mkdir -p build ; \
	for os in darwin linux windows; do \
	for arch in amd64 arm64; do \
	output="build/$(BIN)-v$(GIT_TAG)-$$os-$$arch" ; \
	if [ "$${os}" == "windows" ]; then output="$${output}.exe"; fi ; \
	echo "building: $$output" ; \
	GOOS=$$os GOARCH=$$arch go build -o "$${output}" $(SRC) ; \
	done ; \
	done


build-test-run: build
	cat $(FASTQ) | ./$(BIN)
	gunzip -c $(FASTQGZ) | ./$(BIN)
	./$(BIN) <(gunzip -c $(FASTQGZ))
	./$(BIN) $(FASTQ)
	./$(BIN) $(FASTQGZ)

# # docker build -t stevekm/dump-software-versions:latest .
# DOCKER_TAG:=stevekm/dump-software-versions:$(GIT_TAG)
# docker-build:
# 	docker build -t $(DOCKER_TAG) .

# # docker push stevekm/dump-software-versions:latest
# docker-push:
# 	docker push $(DOCKER_TAG)

# docker-test-run:
# 	docker run --platform linux/amd64 --rm -ti -v ${PWD}:${PWD} --workdir ${PWD} $(DOCKER_TAG) dumpSoftwareVersions -manifestName dump-software-version-demo -manifestVersion 1.0 -nxfVersion 23.04.1 -processLabel CUSTOM_DUMPSOFTWAREVERSIONS example/collated_versions.yml