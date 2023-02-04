ENVVAR=CGO_ENABLED=0 LD_FLAGS=-s
GOOS?=linux
REGISTRY?=kubespace
TAG?=dev
export GOPROXY=https://goproxy.cn,direct

build-binary-amd64: asset-build
	rm -rf bin/amd64
	$(ENVVAR) GOOS=$(GOOS) go build -o bin/amd64/kubespace ./cmd/server
	$(ENVVAR) GOOS=$(GOOS) go build -o bin/amd64/controller-manager ./cmd/controller-manager
	$(ENVVAR) GOOS=$(GOOS) go build -o bin/amd64/kube-agent ./cmd/kube-agent

asset-build: vue-build
	go install github.com/jessevdk/go-assets-builder
	go-assets-builder -s /ui/dist/static ui/dist -o pkg/server/router/assets.go -p router

vue-build:
	cd ui &&\
 	npm config set registry 'https://registry.npm.taobao.org' &&\
 	npm run build &&\
 	cp dist/index.html dist/static/index.html &&\
 	cd ..

docker-builder:
	docker images | grep ospserver-builder || docker build -t ospserver-builder ./builder

build-in-docker: docker-builder
	docker run --rm -v `pwd`:/gopath/src/github.com/kubespace/kubespace ospserver-builder:latest bash -c 'cd /gopath/src/github.com/kubespace/kubespace && make build-binary-amd64'

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