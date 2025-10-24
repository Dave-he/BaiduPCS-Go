package download

import (
	"bufio"
	"errors"
	"io"
	"os"

	"BaiduPCS-Go/baidusdk/utils"
)

func Download(accessToken string, dlink string, outputFilename string) error {
	uri := dlink + "&" + "access_token=" + accessToken
	headers := map[string]string{
		"User-Agent": "pan.baidu.com",
	}

	var postBody io.Reader
	body, statusCode, err := utils.Do2HTTPRequest(uri, postBody, headers)
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return errors.New("download http fail")
	}

	// 下载数据输出到名"outputFilename"的文件
	file, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	write := bufio.NewWriter(file)
	_, err = write.WriteString(body)
	if err != nil {
		return err
	}

	// Flush将缓存的文件真正写入到文件中
	err = write.Flush()
	if err != nil {
		return err
	}

	return nil
}
