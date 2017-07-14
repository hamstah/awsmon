VERSION := $(shell cat ./VERSION)

image:
	docker build -t cirocosta/awsmon .

install:
	go install -v

fmt:
	gofmt -s -w ./main.go
	cd lib && gofmt -s -w .

release:
	git tag -a $(VERSION) -m "Release"
	git push origin $(VERSION)
	goreleaser

.PHONY: install fmt image release
