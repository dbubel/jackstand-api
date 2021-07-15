build:
	go build -ldflags "/
		  -X main.BUILD_GIT_HASH=`git rev-parse HEAD` /
		  -X main.BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%SZ'`" /
		  -v -o jackstand main.go

localstack:
	docker-compose -f docker-compose.dev.yml up -d localstack-s3
	sleep 10
	aws s3 mb s3://jackstand-s3-test --endpoint-url http://localhost:5002

test: localstack
	go fmt ./...
	aws --endpoint-url http://localhost:5002 s3 rm s3://jackstand-s3-test/ --recursive
	go test -v -race -p=1 -cover ./...

lists3:
	aws --endpoint-url http://localhost:5002 s3 ls s3://jackstand-s3-test/ --recursive

latest:
	docker build -t jackstand .
	docker tag jackstand:latest jackstand:latest

start-dev:
	docker-compose -f docker-compose.dev.yml up jackstand nginx

stop-dev:
	docker-compose -f docker-compose.dev.yml down

start:
	docker-compose -f docker-compose.prod.yml up -d jackstand nginx

stop:
	docker-compose -f docker-compose.prod.yml down