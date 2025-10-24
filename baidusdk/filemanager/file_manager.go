package filemanager

import (
	"BaiduPCS-Go/baidusdk/utils"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

// fileManager
//
// RETURNS:
//   - fileManagerReturn: fileManager return
//   - error: the return error if any occurs
func FileManager(accessToken string, arg *FileManagerArg) (FileManagerReturn, error) {
	ret := FileManagerReturn{}

	protocal := "http"
	host := "pan.baidu.com"
	router := "/rest/2.0/xpan/file?method=filemanager&"
	uri := protocal + "://" + host + router

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("opera", arg.Opera)
	uri += params.Encode()

	headers := map[string]string{
		"Host":         host,
		"Content-Type": "application/x-www-form-urlencoded",
	}

	postBody := url.Values{}
	postBody.Add("async", arg.Async)
	fileListJson, err := json.Marshal(arg.FileList)
	if err != nil {
		return ret, err
	}
	postBody.Add("filelist", string(fileListJson))

	body, _, err := utils.DoHTTPRequest(uri, strings.NewReader(postBody.Encode()), headers)
	if err != nil {
		return ret, err
	}
	if err = json.Unmarshal([]byte(body), &ret); err != nil {
		return ret, errors.New("unmarshal filemanager body failed,body")
	}
	if ret.Errno != 0 {
		return ret, errors.New("call filemanager failed")
	}
	return ret, nil
}
