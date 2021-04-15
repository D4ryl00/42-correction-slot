all: install
.PHONY: all

install:
	go install ./...
.PHONY: install

build:
	go build ./...
.PHONY: build

run:
	go run ./...
.PHONY: run

clean:
	rm -f 42-correction-slot
.PHONY: clean
