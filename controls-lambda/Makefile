deploy: compile zip upgrade

BUILD=`git describe --tags --abbrev=0` 
LDFLAGS=-ldflags "-X=main.Version=$(BUILD)"
compile:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -tags lambda.norpc -o bootstrap main.go

xcompile:
	docker run --rm -v "${PWD}":/usr/src/myapp -w /usr/src/myapp -e GOOS=linux -e GOARCH=arm64 golang:1.22 go build $(LDFLAGS) -tags lambda.norpc -o bootstrap main.go

zip:
	zip mac.zip bootstrap

create:
	aws lambda create-function --function-name moneypenny-aws-controls \
	--runtime provided.al2023 --handler bootstrap \
	--architectures arm64 \
	--role ${ROLE} \
	--timeout 60 \
	--environment  "Variables={TIME_ZONE=Europe/Amsterdam,BASIC_USER=${BASIC_USER},BASIC_PASSWORD=${BASIC_PASSWORD}}"
	--zip-file fileb://mac.zip

upgrade:
	aws lambda update-function-code --function-name moneypenny-aws-controls \
	--zip-file fileb://mac.zip

upconfig:
	aws lambda update-function-configuration --function-name moneypenny-aws-controls \
	--timeout 60 \
	--environment  "Variables={TIME_ZONE=Europe/Amsterdam,BASIC_USER=admin,BASIC_PASSWORD=???????}"