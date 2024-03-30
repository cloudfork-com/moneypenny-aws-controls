run:
	sam local start-lambda

test:
	sam local invoke "HandleRequest" -e event.json

deploy: compile zip upgrade

BUILD=`git describe --tags --abbrev=0` 
LDFLAGS=-ldflags "-X=main.Version=$(BUILD)"
compile:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -tags lambda.norpc -o bootstrap main.go

zip:
	zip mac.zip bootstrap

create:
	aws lambda create-function --function-name MAC \
	--runtime provided.al2023 --handler bootstrap \
	--architectures arm64 \
	--role ${ROLE} \
	--timeout 60 \
	--environment  "Variables={TIME_ZONE=Europe/Amsterdam}" \
	--zip-file fileb://mac.zip

upgrade:
	aws lambda update-function-code --function-name MAC \
	--zip-file fileb://mac.zip

upconfig:
	aws lambda update-function-configuration --function-name MAC \
	--environment  "Variables={TIME_ZONE=Europe/Amsterdam}"