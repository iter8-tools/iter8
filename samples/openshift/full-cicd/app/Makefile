all: run

build:
	docker build -t haih/helloworld:v`git rev-parse --short HEAD` .

push:
	docker push haih/helloworld:v`git rev-parse --short HEAD`

run: build push
