ENVVAR=CGO_ENABLED=0 LD_FLAGS=-s
GOOS?=linux
REGISTRY?=openspacee
TAG?=dev


build-binary: clean
	$(ENVVAR) GOOS=$(GOOS) go build -o ospserver

clean:
	rm -f ospserver

docker-builder:
	docker images | grep ospserver-builder || docker build -t ospserver-builder ./builder

vue-build:
	cd ui && npm run build && cp dist/index.html dist/static/index.html

asset-build:
	go-assets-builder -s /ui/dist/static ui/dist -o pkg/router/assets.go -p router

build-in-docker: clean docker-builder
	docker run --rm -v `pwd`:/gopath/src/github.com/openspacee/osp/ ospserver-builder:latest bash -c 'cd /gopath/src/github.com/openspacee/osp && make build-binary'

make-base-image:
	docker images | grep openspacee/ospserver-base || docker build -t openspacee/ospserver-base ./base-image

make-image: vue-build asset-build build-in-docker make-base-image
	docker build -t ${REGISTRY}/osp:${TAG} .

push-image:
	docker push ${REGISTRY}/osp:${TAG}
	docker tag ${REGISTRY}/osp:${TAG} ${REGISTRY}/osp:latest
	docker push ${REGISTRY}/osp:latest

execute-release: make-image push-image

# make build-in-docker
# TAG=v1.2.0  make make-image
# TAG=v1.3.0 make execute-release