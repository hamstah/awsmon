image:
	docker build -t cirocosta/awsmon .

install:
	cd ./awsmon && go install -v

build:
	cd ./awsmon && go build -v

fmt:
	cd ./awsmon && gofmt -s -w .

.PHONY: install build fmt image
