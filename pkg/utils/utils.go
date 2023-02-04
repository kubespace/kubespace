package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Void struct{}

var Ok Void

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

func HttpPost(url string, body interface{}) ([]byte, error) {
	bodyBytes, _ := json.Marshal(body)
	klog.Infof("request for url=%s", url)
	klog.Infof("request body: %s", string(bodyBytes))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	data, errReadBody := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if errReadBody != nil {
		klog.Error("read received http resp body error: error=", err)
		return nil, err
	}
	klog.Infof("doRequest get response: url=%s, status=%v", url, resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		klog.Errorf("receive http code not 200: httpcode=%d, data=%s", resp.StatusCode, string(data))
		return data, fmt.Errorf("status code %v", resp.StatusCode)
	} else {
		return data, nil
	}
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
