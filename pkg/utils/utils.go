package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func Contains(strList []string, str string) bool {
	for _, s := range strList {
		if s == str {
			return true
		}
	}
	return false
}

func ParseBool(str string) bool {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	case "0", "f", "F", "false", "FALSE", "False":
		return false
	}
	return false
}

func CreateUUID() string {
	return uuid.New().String()
}

func ShortUUID() string {
	return shortuuid.New()
}

func StringNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Encrypt(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyEmailFormat(email string) bool {
	//pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// VerifyMobileFormat mobile verify
func VerifyMobileFormat(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func MergeMap(mObj ...map[string]interface{}) map[string]interface{} {
	newObj := map[string]interface{}{}
	for _, m := range mObj {
		for k, v := range m {
			if _, ok := newObj[k]; ok {
				newObj[k] = fmt.Sprintf("%v,%v", newObj[k], v)
			} else {
				newObj[k] = v
			}
		}
	}
	return newObj
}

func MergeReplaceMap(mObj ...map[string]interface{}) map[string]interface{} {
	newObj := map[string]interface{}{}
	for _, m := range mObj {
		for k, v := range m {
			newObj[k] = v
		}
	}
	return newObj
}

func ConvertTypeByJson(srcObj interface{}, destPtr interface{}) (err error) {
	srcBytes, ok := srcObj.([]byte)
	if !ok {
		srcBytes, err = json.Marshal(srcObj)
		if err != nil {
			return
		}
	}
	return json.Unmarshal(srcBytes, destPtr)
}

func HandleCrash(additionalHandlers ...func(interface{})) {
	if r := recover(); r != nil {
		if r == http.ErrAbortHandler {
			// honor the http.ErrAbortHandler sentinel panic value:
			//   ErrAbortHandler is a sentinel panic value to abort a handler.
			//   While any panic from ServeHTTP aborts the response to the client,
			//   panicking with ErrAbortHandler also suppresses logging of a stack trace to the server's error log.
			return
		}

		// Same as stdlib http server code. Manually allocate stack trace buffer size
		// to prevent excessively large logs
		const size = 64 << 10
		stacktrace := make([]byte, size)
		stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
		if _, ok := r.(string); ok {
			klog.Errorf("Observed a panic: %s\n%s", r, stacktrace)
		} else {
			klog.Errorf("Observed a panic: %#v (%v)\n%s", r, r, stacktrace)
		}
		for _, fn := range additionalHandlers {
			fn(r)
		}
	}
}

func StringPtr(s string) *string { return &s }

// GetCodeRepoName 获取代码库的项目名
// 如：https://github.com/test/testrepo.git -> testrepo
// git@github.com/test/testrepo.git -> testrepo
func GetCodeRepoName(codeUrl string) string {
	codeSplit := strings.Split(codeUrl, "/")
	codeDir := codeSplit[len(codeSplit)-1]
	codeSplit = strings.Split(codeDir, ".")
	return codeSplit[0]
}

// GetImageName 获取镜像的名称
// 如：docker.io/kubespace/kubespace:latest -> kubespace/kubespace
func GetImageName(img string) string {
	_, name, _ := ParseImageName(img, true)
	return name
}

// ParseImageName 获取镜像的名称
// 如：docker.io/kubespace/kubespace:latest -> docker.io kubespace/kubespace latest
func ParseImageName(img string, withDefault bool) (string, string, string) {
	var registry, name, tag string
	if withDefault {
		registry = "docker.io"
		tag = "latest"
	}
	splitImg := strings.Split(img, ":")
	if len(splitImg) == 1 {
		name = img
	} else if len(splitImg) == 2 {
		if strings.Contains(splitImg[0], "/") {
			name = splitImg[0]
			tag = splitImg[1]
		} else {
			// 127.0.0.1:5000/busybox
			name = img
		}
	} else {
		// 去掉tag
		name = strings.Join(splitImg[0:len(splitImg)-1], ":")
		tag = splitImg[len(splitImg)-1]
	}
	if strings.Contains(strings.Split(name, "/")[0], ".") {
		name = strings.Join(strings.Split(name, "/")[1:], "/")
		registry = strings.Split(name, "/")[0]
	}
	return registry, name, tag
}

// RequestHost 获取http请求的server host
func RequestHost(r *http.Request) (host string) {
	switch {
	case r.Header.Get("X-Host") != "":
		return r.Header.Get("X-Host")
	case r.Host != "":
		return r.Host
	case r.URL.Host != "":
		return r.URL.Host
	default:
		return r.Host
	}
}

func ParseUint(s string) (uint, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	return uint(i), err
}

func GetMapStringValue(m map[string]interface{}, k string) (s string, ok bool) {
	v, ok := m[k]
	if !ok {
		return
	}
	s, ok = v.(string)
	return
}
