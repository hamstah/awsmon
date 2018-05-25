VERSION := $(shell cat ./VERSION)

install:
	go install -v

image:
	docker build -t cirocosta/awsmon .

fmt:
	go fmt ./...

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist

.PHONY: install fmt image release
