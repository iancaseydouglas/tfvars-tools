.PHONY: all build test clean docker-build docker-test docker-run docker-shell lint fmt pr-create pr-check dev-start dev-stop dev-exec deps-update docs release

IMAGE_NAME := tfvars-tools
CONTAINER_NAME := tfvars-tools-dev
VERSION := $(shell git describe --tags --always)

all: docker-build docker-test

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION)-a .

docker-test:
	docker run --rm $(IMAGE_NAME:$(VERSION)) go test ./...

build:
	docker run $(IMAGE_NAME):$(VERSION)-a go build -o bin/findtfvars ./cmd/findtfvars
	docker run $(IMAGE_NAME):$(VERSION)-a go build -o bin/tfvars-to-args ./cmd/tfvars-to-args
	# docker run --rm -v $(PWD):/app -w /app $(IMAGE_NAME):$(VERSION) go build -o bin/findtfvars ./cmd/findtfvars
	# docker run --rm -v $(PWD):/app -w /app $(IMAGE_NAME):$(VERSION) go build -o bin/tfvars-to-args ./cmd/tfvars-to-args

docker-shell:
	docker run --rm -it -v $(PWD):/app -w /app $(IMAGE_NAME):$(VERSION)-a /bin/bash

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.55 golangci-lint run

fmt:
	docker run --rm -v $(PWD):/app -w /app $(IMAGE_NAME) go fmt ./...

pr-create:
	@read -p "Enter feature branch name: " branch_name; 	git checkout -b $$branch_name; 	git push -u origin $$branch_name

pr-check: docker-build docker-test lint fmt
	@echo "All checks passed. Ready to create a pull request."

clean:
	rm -rf bin
	docker rm -f $(CONTAINER_NAME) 2>/dev/null || true

dev-start:
	docker-compose up -d

dev-stop:
	docker-compose down

dev-exec:
	@if [ -z "$(CMD)" ]; then 		echo "Usage: make dev-exec CMD='your command'"; 	else 		docker-compose exec dev $(CMD); 	fi

deps-update:
	docker run --rm -v $(PWD):/app -w /app $(IMAGE_NAME) go get -u ./...
	docker run --rm -v $(PWD):/app -w /app $(IMAGE_NAME) go mod tidy

docs:
	docker run --rm -v $(PWD):/app -w /app $(IMAGE_NAME) go doc -all > DOCS.md

release:
	@echo "Creating release $(VERSION)"
	@go mod tidy
	@git tag $(VERSION)
	@git push origin $(VERSION)
	@goreleaser release --rm-dist
