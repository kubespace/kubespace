cd ui
npm config set registry https://registry.npm.taobao.org
npm install
npm run build
cp dist/index.html dist/static/index.html
cd ..
export GOPROXY=https://proxy.golang.com.cn,direct
go get github.com/jessevdk/go-assets-builder
go-assets-builder -s /ui/dist/static ui/dist -o pkg/router/assets.go -p router
make build-binary
