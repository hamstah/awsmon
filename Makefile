VERSION := $(shell cat ./VERSION)

image:
	docker build -t cirocosta/awsmon .

install:
	go install -v

fmt:
	gofmt -s -w ./main.go
	cd lib && gofmt -s -w .

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist

.PHONY: install fmt image release
