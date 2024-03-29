run:
	sam local start-lambda

test:
	sam local invoke "HandleRequest" -e event.json

deploy: compile zip upgrade

compile:
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap main.go

zip:
	zip mac.zip bootstrap

create:
	aws lambda create-function --function-name MAC \
	--runtime provided.al2023 --handler bootstrap \
	--architectures arm64 \
	--role arn:aws:iam::413736178351:role/moneypenny-aws-controls \
	--zip-file fileb://mac.zip

upgrade:
	aws lambda update-function-code --function-name MAC \
	--zip-file fileb://mac.zip