package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/iikira/iikira-go-utils/requester"
)

// HTTPClient 统一的HTTP客户端接口
type HTTPClient interface {
	Get(url string) (*http.Response, error)
	Post(url string, data interface{}) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient 默认HTTP客户端实现
type DefaultHTTPClient struct {
	client *requester.HTTPClient
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: requester.NewHTTPClient(),
	}
}

// Get 执行GET请求
func (c *DefaultHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Req(http.MethodGet, url, nil, nil)
}

// Post 执行POST请求
func (c *DefaultHTTPClient) Post(url string, data interface{}) (*http.Response, error) {
	return c.client.Req(http.MethodPost, url, data, nil)
}

// Do 执行自定义请求
func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// SetTimeout 设置超时时间
func (c *DefaultHTTPClient) SetTimeout(timeout time.Duration) {
	c.client.SetTimeout(timeout)
}

// SetUserAgent 设置User-Agent
func (c *DefaultHTTPClient) SetUserAgent(ua string) {
	c.client.SetUserAgent(ua)
}

// ErrorHandler 统一的错误处理接口
type ErrorHandler interface {
	HandleError(operation string, err error) error
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct{}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(operation string, err error) error {
	return fmt.Errorf("%s失败: %v", operation, err)
}

// RetryOperation 重试操作
func RetryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		
		if i < maxRetries {
			fmt.Printf("⚠️  操作失败，%v 后重试 (%d/%d): %v\n", delay, i+1, maxRetries, err)
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, err)
}

// SafeCreateFile 安全创建文件
func SafeCreateFile(filePath string) (*os.File, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}
	
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %v", err)
	}
	
	return file, nil
}

// ValidateFileSize 验证文件大小
func ValidateFileSize(filePath string, expectedSize uint64) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}
	
	actualSize := uint64(stat.Size())
	if actualSize != expectedSize {
		return fmt.Errorf("文件大小不匹配: 期望 %d 字节，实际 %d 字节", expectedSize, actualSize)
	}
	
	return nil
}

// FormatFileSize 格式化文件大小
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

// CopyWithProgress 带进度的文件复制
func CopyWithProgress(dst io.Writer, src io.Reader, size int64, description string) (int64, error) {
	// 这里可以集成进度条逻辑
	return io.Copy(dst, src)
}

// EnsureDir 确保目录存在
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}