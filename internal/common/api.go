package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
)

// APIResponse 通用API响应结构
type APIResponse struct {
	Errno     int         `json:"errno"`
	Errmsg    string      `json:"errmsg"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
}

// APIClient 通用API客户端
type APIClient struct {
	BaseURL    string
	HTTPClient HTTPClient
	Headers    map[string]string
}

// NewAPIClient 创建新的API客户端
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: NewHTTPClient(),
		Headers:    make(map[string]string),
	}
}

// SetHeader 设置请求头
func (c *APIClient) SetHeader(key, value string) {
	c.Headers[key] = value
}

// Get 执行GET请求
func (c *APIClient) Get(endpoint string, params map[string]string) (*APIResponse, error) {
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %v", err)
	}

	// 添加查询参数
	query := u.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

// Post 执行POST请求
func (c *APIClient) Post(endpoint string, data interface{}) (*APIResponse, error) {
	// 实现POST请求逻辑
	return nil, fmt.Errorf("POST方法待实现")
}

// parseResponse 解析API响应
func (c *APIClient) parseResponse(resp *http.Response) (*APIResponse, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if apiResp.Errno != 0 {
		return &apiResp, fmt.Errorf("API错误: %d - %s", apiResp.Errno, apiResp.Errmsg)
	}

	return &apiResp, nil
}

// Config 通用配置结构
type Config struct {
	AppKey      string `json:"app_key"`
	SecretKey   string `json:"secret_key"`
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	BDUSS       string `json:"bduss"`
	STOKEN      string `json:"stoken"`
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	AppKey:    "iYCeC9g08h5vuP9UqvPHKKSVrKFXGa1v",
	SecretKey: "jXiFMOPVPCWlO2M5CwWQzffpNPaGTRBG",
	BDUSS:     "nZidms5WG1IamlERkRJZXplTmdoUGNoRlFxcUR1UHR4V3ZBSkJlZUNxVUJkQ0ZwRVFBQUFBJCQAAAAAAAAAAAEAAADybtIuMTA2NDA0MjQxMWwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAHn-WgB5~loaSTOKEN:40e4dd9e6d04e08439490bed5e5365efce212fec312933f467ef58acf7f83874",
}

// LoadConfig 加载配置
func LoadConfig(path string) (*Config, error) {
	if !FileExists(path) {
		return DefaultConfig, nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// SaveConfig 保存配置
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}