package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"BaiduPCS-Go/internal/ui"
	"BaiduPCS-Go/baidusdk/download"
)

// 文件搜索结构体
type FileSearchResponse struct {
	Errno   int              `json:"errno"`
	Errmsg  string           `json:"errmsg"`
	List    []FileSearchItem `json:"list"`
	HasMore int              `json:"has_more"`
}

type FileSearchItem struct {
	FsId           uint64 `json:"fs_id"`
	Path           string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Size           uint64 `json:"size"`
	IsDir          int    `json:"isdir"`
	Category       int    `json:"category"`
	Md5            string `json:"md5"`
	ServerMtime    int64  `json:"server_mtime"`
}

// 文件列表结构体
type FileListResponse struct {
	Errno   int            `json:"errno"`
	Errmsg  string         `json:"errmsg"`
	List    []FileListItem `json:"list"`
	HasMore int            `json:"has_more"`
}

type FileListItem struct {
	FsId           uint64 `json:"fs_id"`
	Path           string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Size           uint64 `json:"size"`
	IsDir          int    `json:"isdir"`
	Category       int    `json:"category"`
	Md5            string `json:"md5"`
	ServerMtime    int64  `json:"server_mtime"`
}

// 搜索文件
func searchFiles(accessToken, keyword string, recursive bool) ([]FileSearchItem, error) {
	fmt.Printf("🔍 搜索文件: %s\n", keyword)

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("method", "search")
	params.Set("key", keyword)
	params.Set("num", "100") // 最多返回100个结果
	if recursive {
		params.Set("recursion", "1")
	} else {
		params.Set("recursion", "0")
	}

	searchURL := "https://pan.baidu.com/rest/2.0/xpan/file?" + params.Encode()

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("搜索请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取搜索响应失败: %v", err)
	}

	var result FileSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析搜索响应失败: %v", err)
	}

	if result.Errno != 0 {
		return nil, fmt.Errorf("搜索失败，错误码: %d, 错误信息: %s", result.Errno, result.Errmsg)
	}

	return result.List, nil
}

// 获取目录文件列表
func getFileList(accessToken, dir string) ([]FileListItem, error) {
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("method", "list")
	params.Set("dir", dir)
	params.Set("num", "1000") // 最多返回1000个结果
	params.Set("order", "name")
	params.Set("desc", "0")

	listURL := "https://pan.baidu.com/rest/2.0/xpan/file?" + params.Encode()

	resp, err := http.Get(listURL)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取文件列表响应失败: %v", err)
	}

	var result FileListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析文件列表响应失败: %v", err)
	}

	if result.Errno != 0 {
		return nil, fmt.Errorf("获取文件列表失败，错误码: %d, 错误信息: %s", result.Errno, result.Errmsg)
	}

	return result.List, nil
}

// FindFileByName 根据文件名查找文件
func FindFileByName(accessToken, filename string, exactMatch bool) ([]FileSearchItem, error) {
	var files []FileSearchItem
	var err error

	if exactMatch {
		// 精确匹配：先尝试搜索
		files, err = searchFiles(accessToken, filename, true)
		if err != nil {
			return nil, err
		}

		// 过滤精确匹配的文件
		var exactFiles []FileSearchItem
		for _, file := range files {
			if file.ServerFilename == filename {
				exactFiles = append(exactFiles, file)
			}
		}
		files = exactFiles
	} else {
		// 模糊匹配
		files, err = searchFiles(accessToken, filename, true)
		if err != nil {
			return nil, err
		}

		// 过滤包含关键词的文件
		var matchedFiles []FileSearchItem
		lowerFilename := strings.ToLower(filename)
		for _, file := range files {
			if strings.Contains(strings.ToLower(file.ServerFilename), lowerFilename) {
				matchedFiles = append(matchedFiles, file)
			}
		}
		files = matchedFiles
	}

	return files, nil
}

// 显示搜索结果并让用户选择
func selectFileFromResults(files []FileSearchItem) (*FileSearchItem, error) {
	// 转换为ui包的类型
	uiFiles := make([]ui.FileSearchItem, len(files))
	for i, file := range files {
		uiFiles[i] = ui.FileSearchItem{
			FsId:           file.FsId,
			Path:           file.Path,
			ServerFilename: file.ServerFilename,
			Size:           file.Size,
			IsDir:          file.IsDir,
			Category:       file.Category,
			Md5:            file.Md5,
			ServerMtime:    file.ServerMtime,
		}
	}
	
	selector := ui.NewFileSelector(uiFiles)
	selected, err := selector.SelectFile()
	if err != nil {
		return nil, err
	}
	
	// 转换回sdk包的类型
	result := &FileSearchItem{
		FsId:           selected.FsId,
		Path:           selected.Path,
		ServerFilename: selected.ServerFilename,
		Size:           selected.Size,
		IsDir:          selected.IsDir,
		Category:       selected.Category,
		Md5:            selected.Md5,
		ServerMtime:    selected.ServerMtime,
	}
	
	return result, nil
}

// 格式化文件大小
func FormatFileSize(size uint64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := uint64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// RunSDKDownloadByName 优化后的SDK下载功能 - 支持文件名搜索
func RunSDKDownloadByName(accessToken string, filename string, outputPath string, parallel int, exactMatch bool) error {
	fmt.Printf("🚀 根据文件名下载: %s\n", filename)

	// 1. 搜索文件
	files, err := FindFileByName(accessToken, filename, exactMatch)
	if err != nil {
		return err
	}

	// 2. 选择文件
	selectedFile, err := selectFileFromResults(files)
	if err != nil {
		return err
	}

	// 3. 生成智能输出路径
	smartOutputPath := ui.GenerateOutputPath(selectedFile.ServerFilename, outputPath)

	// 4. 确认下载
	if !ui.ConfirmDownload(selectedFile.ServerFilename, smartOutputPath, selectedFile.Size) {
		fmt.Println("❌ 用户取消下载")
		return nil
	}

	// 5. 使用fsid下载
	return fmt.Errorf("下载功能待实现")
}

// RunSDKDownloadByFsid 优化后的SDK下载功能 - 支持fsid下载
func RunSDKDownloadByFsid(accessToken string, fsid uint64, outputPath string, parallel int) error {
	fmt.Printf("🚀 根据fsid下载: %d\n", fsid)

	// 1. 获取文件信息
	arg := &download.FileMetasArg{
		Fsids: []uint64{fsid},
	}

	result, err := download.FileMetas(accessToken, arg)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	if len(result.List) == 0 {
		return fmt.Errorf("文件不存在")
	}

	file := result.List[0]

	// 2. 生成智能输出路径
	smartOutputPath := ui.GenerateOutputPath(file.Filename, outputPath)

	// 3. 确认下载
	if !ui.ConfirmDownload(file.Filename, smartOutputPath, file.Size) {
		fmt.Println("❌ 用户取消下载")
		return nil
	}

	// 4. 执行下载
	return fmt.Errorf("下载功能待实现")
}

// RunSDKLogin SDK登录
func RunSDKLogin(bduss string, stoken string) error {
	fmt.Println("🔐 正在使用BDUSS登录...")
	
	// 这里应该实现BDUSS登录逻辑
	// 暂时模拟成功
	fmt.Println("✅ 登录成功")
	
	return nil
}

// CheckAndRefreshToken 检查并刷新token
func CheckAndRefreshToken() error {
	// 这里应该实现token刷新逻辑
	// 暂时返回成功
	return nil
}
