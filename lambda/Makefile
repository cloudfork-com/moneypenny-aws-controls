run:
	sam local start-lambda

test:
	sam local invoke "HandleRequest" -e event.json

deploy: compile zip create

compile:
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go

zip:
	zip mac.zip bootstrap

create:
	aws lambda create-function --function-name MAC \
	--runtime provided.al2023 --handler bootstrap \
	--architectures arm64 \
	--role arn:aws:iam::111122223333:role/lambda-ex \
	--zip-file fileb://mac.zip