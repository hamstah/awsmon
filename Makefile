image:
	docker build -t cirocosta/awsmon .

install:
	go install -v

fmt:
	gofmt -s -w ./main.go
	cd lib && gofmt -s -w .

release:
	goreleaser --snapshot


.PHONY: install fmt image
