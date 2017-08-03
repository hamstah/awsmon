VERSION := $(shell cat ./VERSION)

image:
	docker build -t cirocosta/awsmon .

install:
	go install -v

fmt:
	go fmt ./main.go
	cd lib && go fmt

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist

.PHONY: install fmt image release
