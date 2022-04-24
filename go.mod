module github.com/kubespace/kubespace

go 1.16

require (
	github.com/gin-gonic/gin v1.7.7
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-redis/redis/v8 v8.11.4
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	gorm.io/driver/mysql v1.2.1
	gorm.io/gorm v1.22.4
	helm.sh/helm/v3 v3.7.2
	k8s.io/api v0.22.4
	k8s.io/apimachinery v0.22.4
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.3.0
)
