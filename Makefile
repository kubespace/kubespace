ENVVAR=CGO_ENABLED=0 LD_FLAGS=-s
GOOS?=linux
REGISTRY?=kubespace
TAG?=dev
export GOPROXY=https://goproxy.cn,direct

build-binary: asset-build
	rm -rf bin/amd64
	$(ENVVAR) GOOS=$(GOOS) go build -o bin/amd64/kubespace-server ./cmd/server
	$(ENVVAR) GOOS=$(GOOS) go build -o bin/amd64/controller-manager ./cmd/controller-manager
	$(ENVVAR) GOOS=$(GOOS) go build -o bin/amd64/kube-agent ./cmd/kube-agent
	$(ENVVAR) GOOS=$(GOOS) go build -o bin/amd64/spacelet ./cmd/spacelet

	rm -rf assets
	mkdir -p assets/linux/amd64
	cp bin/amd64/spacelet assets/linux/amd64/spacelet

ifdef BUILD_ARM64
	rm -rf bin/arm64
	$(ENVVAR) GOOS=$(GOOS) GOARCH=arm64 go build -o bin/arm64/kubespace-server ./cmd/server
	$(ENVVAR) GOOS=$(GOOS) GOARCH=arm64 go build -o bin/arm64/controller-manager ./cmd/controller-manager
	$(ENVVAR) GOOS=$(GOOS) GOARCH=arm64 go build -o bin/arm64/kube-agent ./cmd/kube-agent
	$(ENVVAR) GOOS=$(GOOS) GOARCH=arm64 go build -o bin/arm64/spacelet ./cmd/spacelet

	mkdir -p assets/linux/arm64
	cp bin/arm64/spacelet assets/linux/arm64/spacelet
endif

asset-build: vue-build
	go install github.com/jessevdk/go-assets-builder@latest
	go-assets-builder -s /ui/dist/static ui/dist -o pkg/server/router/assets.go -p router

vue-build:
	cd ui &&\
	npm config set registry 'https://registry.npm.taobao.org' &&\
	npm install &&\
	npm run build &&\
	cp dist/index.html dist/static/index.html &&\
	cd ..
