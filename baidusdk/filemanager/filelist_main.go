package filemanager

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type FilelistArg struct {
	Dir   string `json:"dir"`
	Order string `json:"order"`
	Start uint64 `json:"start"`
	Limit uint64 `json:"limit"`
	Desc  int    `json:"desc"`
}

func NewFilelistArg(dir string, order string, start uint64, limit uint64, desc int) *FilelistArg {
	s := new(FilelistArg)
	s.Dir = dir
	s.Start = start
	s.Limit = limit
	s.Order = order
	s.Desc = desc
	return s
}

type FileInfo struct {
	Fsid           uint64            `json:"fs_id"`
	Path           string            `json:"path"`
	ServerFilename string            `json:"server_filename"`
	Isdir          int               `json:"isdir"`
	Size           uint64            `json:"size"`
	Category       int               `json:"category"`
	Md5            string            `json:"md5"`
	DirEmpty       int               `json:"dir_empty"`
	LocalCtime     uint64            `json:"local_ctime"`
	LocalMtime     uint64            `json:"local_mtime"`
	ServerCtime    uint64            `json:"server_ctime"`
	ServerMtime    uint64            `json:"server_mtime"`
	Thumbs         map[string]string `json:"thumbs"` // 当文件类型为图片时，且请求参数含有web=1时，返回thumbs
}

type FilelistReturn struct {
	Errno     int        `json:"errno"`
	RequestID int        `json:"request_id"`
	Guid      int        `json:"guid"`
	GuidInfo  string     `json:"guid_info"`
	List      []FileInfo `json:"list"` //文件列表
}

// filelist
//
// RETURNS:
//     - filelistReturn: filelist return
//     - error: the return error if any occurs
func Filelist(accessToken string, arg *FilelistArg) (FilelistReturn, error) {
	ret := FilelistReturn{}

	protocal := "https"
	host := "pan.baidu.com"
	router := "/rest/2.0/xpan/file?method=list&"
	uri := protocal + "://" + host + router

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("dir", arg.Dir)
	params.Set("start", strconv.FormatUint(arg.Start, 10))
	params.Set("limit", strconv.FormatUint(arg.Limit, 10))
	params.Set("order", arg.Order)
	params.Set("desc", strconv.FormatInt(int64(arg.Desc), 10))
	params.Set("folder", "0")
	params.Set("web", "1")
	params.Set("show_empty", "1")
	uri += params.Encode()

	headers := map[string]string{
		"Host":         host,
		"Content-Type": "application/x-www-form-urlencoded",
	}

	var postBody io.Reader
	body, _, err := DoHTTPRequest(uri, postBody, headers)
	if err != nil {
		return ret, err
	}
	if err = json.Unmarshal([]byte(body), &ret); err != nil {
		return ret, errors.New("unmarshal filemlist body failed,body")
	}
	if ret.Errno != 0 {
		return ret, errors.New("call filelist failed")
	}
	return ret, nil
}

//DoHTTPRequest
//
// RETURNS:
//     - filelistReturn: respBody return
//     - filelistReturn: status code return
//     - error: the return error if any occurs
func DoHTTPRequest(url string, body io.Reader, headers map[string]string) (string, int, error) {
	timeout := 5 * time.Second
	retryTimes := 3
	tr := &http.Transport{
		MaxIdleConnsPerHost: -1,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	httpClient.Timeout = timeout
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", 0, err
	}
	// request header
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	var resp *http.Response
	for i := 1; i <= retryTimes; i++ {
		resp, err = httpClient.Do(req)
		if err == nil {
			break
		}
		if i == retryTimes {
			return "", 0, err
		}
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}
	return string(respBody), resp.StatusCode, nil
}

func main() {

	// 使用示例

	// 用户的access_token
	accessToken := "your-access-token"

	// 需要查list的目录
	// 必须为目录
	// 查"/清华计算机/清华大学计算机系网络课程"下的文件
	dir := "/清华计算机/清华大学计算机系网络课程"
	dir = "/电影/老Y记/SE01/"

	// start起始位置，limit查询数目
	// 示例为从0开始查2条
	var start uint64
	var limit uint64
	start = 0
	limit = 2

	// 示例为按照文件修改时间，降序排列文件
	// 排序字段
	order := "time"
	// 降序
	desc := 1

	// call API
	arg := NewFilelistArg(dir, order, start, limit, desc)
	if ret, err := Filelist(accessToken, arg); err != nil {
		fmt.Printf("[msg: file list error] [err:%v]", err.Error())
	} else {
		fmt.Printf("ret:%+v", ret)
		fmt.Println()
		fmt.Printf("file list:%+v", ret.List)
	}
}
