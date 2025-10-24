package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"strconv"

	"BaiduPCS-Go/baidusdk/utils"
)

func Upload(accessToken string, arg *UploadArg) (UploadReturn, error) {
	ret := UploadReturn{}

	// 打开文件句柄操作
	fileHandle, err := os.Open(arg.LocalFile)
	if err != nil {
		return ret, errors.New("open local file failed: " + err.Error())
	}
	defer fileHandle.Close()

	// 获取文件当前信息
	fileInfo, err := fileHandle.Stat()
	if err != nil {
		return ret, err
	}

	// 读取文件块 - 优化内存使用，避免一次性读取整个文件
	const maxChunkSize = 4 * 1024 * 1024 // 4MB chunks
	fileSize := fileInfo.Size()
	chunkSize := maxChunkSize
	if fileSize < maxChunkSize {
		chunkSize = int(fileSize)
	}

	buf := make([]byte, chunkSize)
	n, err := fileHandle.Read(buf)
	if err != nil && err != io.EOF {
		return ret, err
	}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "file")
	if err != nil {
		return ret, err
	}

	//iocopy
	_, err = io.Copy(fileWriter, bytes.NewReader(buf[0:n]))
	if err != nil {
		return ret, err
	}

	bodyWriter.Close()

	protocal := "https"
	host := "d.pcs.baidu.com"
	router := "/rest/2.0/pcs/superfile2?method=upload&"
	uri := protocal + "://" + host + router

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("path", arg.Path)
	params.Set("uploadid", arg.UploadId)
	params.Set("partseq", strconv.Itoa(arg.Partseq))
	uri += params.Encode()

	contentType := bodyWriter.FormDataContentType()
	headers := map[string]string{
		"Host":         host,
		"Content-Type": contentType,
	}

	body, _, err := utils.SendHTTPRequest(uri, bodyBuf, headers)
	if err != nil {
		return ret, err
	}
	if err := json.Unmarshal([]byte(body), &ret); err != nil {
		return ret, errors.New("unmarshal response body failed: " + err.Error())
	}

	if ret.Md5 == "" {
		return ret, errors.New("upload failed: md5 is empty in response")
	}
	return ret, nil
}
