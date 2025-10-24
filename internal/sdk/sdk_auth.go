package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"BaiduPCS-Go/internal/pcsconfig"
)

// SDK默认配置
const (
	DefaultAppKey    = "t01UV5SNjSyo3uI2HyIbwB6Agy01wrtg"
	DefaultSecretKey = "Z3SI78r1mId9Mx77aC3wFb66wXwVAOTY"
	DefaultSignKey   = "T0^oa5fM6@WHTOknJx8PUbpJEkeAl1Ew"
)

// OAuth相关结构体
type OAuthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	SessionKey       string `json:"session_key"`
	SessionSecret    string `json:"session_secret"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type DeviceCodeResponse struct {
	DeviceCode       string `json:"device_code"`
	UserCode         string `json:"user_code"`
	VerificationUrl  string `json:"verification_url"`
	QrCodeUrl        string `json:"qrcode_url"`
	ExpiresIn        int    `json:"expires_in"`
	Interval         int    `json:"interval"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// 生成签名
func generateSign(params map[string]string, secretKey string) string {
	// 按key排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}
	signStr.WriteString(secretKey)

	// MD5加密
	hash := md5.Sum([]byte(signStr.String()))
	return hex.EncodeToString(hash[:])
}

// 获取设备码
func getDeviceCode(appKey string) (*DeviceCodeResponse, error) {
	fmt.Println("🔐 获取设备授权码...")

	params := map[string]string{
		"response_type": "device_code",
		"client_id":     appKey,
		"scope":         "basic,netdisk",
	}

	// 构建请求
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}

	resp, err := http.PostForm("https://openapi.baidu.com/oauth/2.0/device/code", data)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result DeviceCodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取设备码失败: %s - %s", result.Error, result.ErrorDescription)
	}

	return &result, nil
}

// 轮询获取访问令牌
func pollAccessToken(appKey, secretKey, deviceCode string, interval int) (*OAuthTokenResponse, error) {
	fmt.Println("⏳ 等待用户授权...")

	for {
		time.Sleep(time.Duration(interval) * time.Second)

		params := map[string]string{
			"grant_type":    "device_token",
			"code":          deviceCode,
			"client_id":     appKey,
			"client_secret": secretKey,
		}

		data := url.Values{}
		for k, v := range params {
			data.Set(k, v)
		}

		resp, err := http.PostForm("https://openapi.baidu.com/oauth/2.0/token", data)
		if err != nil {
			fmt.Printf("⚠️  请求失败: %v，继续等待...\n", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("⚠️  读取响应失败: %v，继续等待...\n", err)
			continue
		}

		var result OAuthTokenResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("⚠️  解析响应失败: %v，继续等待...\n", err)
			continue
		}

		if result.Error != "" {
			if result.Error == "authorization_pending" {
				fmt.Print(".")
				continue
			} else if result.Error == "slow_down" {
				interval += 5
				fmt.Printf("⚠️  请求过于频繁，增加等待时间到 %d 秒\n", interval)
				continue
			} else if result.Error == "expired_token" {
				return nil, fmt.Errorf("设备码已过期，请重新获取")
			} else {
				return nil, fmt.Errorf("获取访问令牌失败: %s - %s", result.Error, result.ErrorDescription)
			}
		}

		return &result, nil
	}
}

// 刷新访问令牌
func refreshAccessToken(appKey, secretKey, refreshToken string) (*OAuthTokenResponse, error) {
	fmt.Println("🔄 刷新访问令牌...")

	params := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     appKey,
		"client_secret": secretKey,
	}

	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}

	resp, err := http.PostForm("https://openapi.baidu.com/oauth/2.0/token", data)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result OAuthTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("刷新令牌失败: %s - %s", result.Error, result.ErrorDescription)
	}

	return &result, nil
}

// SDK登录流程
func runSDKLogin(appKey, secretKey string) error {
	if appKey == "" {
		appKey = DefaultAppKey
	}
	if secretKey == "" {
		secretKey = DefaultSecretKey
	}

	fmt.Println("🚀 开始SDK登录流程...")
	fmt.Printf("📱 AppKey: %s\n", appKey)

	// 1. 获取设备码
	deviceResp, err := getDeviceCode(appKey)
	if err != nil {
		return err
	}

	fmt.Println("\n📋 请按以下步骤完成授权:")
	fmt.Printf("1️⃣  打开浏览器访问: %s\n", deviceResp.VerificationUrl)
	fmt.Printf("2️⃣  输入用户码: %s\n", deviceResp.UserCode)
	fmt.Printf("3️⃣  或者扫描二维码: %s\n", deviceResp.QrCodeUrl)
	fmt.Printf("⏰ 授权码有效期: %d 秒\n", deviceResp.ExpiresIn)
	fmt.Println("\n⏳ 等待授权完成...")

	// 2. 轮询获取访问令牌
	tokenResp, err := pollAccessToken(appKey, secretKey, deviceResp.DeviceCode, deviceResp.Interval)
	if err != nil {
		return err
	}

	fmt.Println("\n✅ 授权成功!")
	fmt.Printf("🔑 AccessToken: %s\n", tokenResp.AccessToken)
	fmt.Printf("🔄 RefreshToken: %s\n", tokenResp.RefreshToken)
	fmt.Printf("⏰ 有效期: %d 秒\n", tokenResp.ExpiresIn)
	fmt.Printf("📋 权限范围: %s\n", tokenResp.Scope)

	// 3. 保存到配置
	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser.UID == 0 {
		fmt.Println("⚠️  未登录百度账号，将创建新的配置项")
		// 这里可以创建一个新的用户配置
	}

	activeUser.AccessToken = tokenResp.AccessToken
	activeUser.RefreshToken = tokenResp.RefreshToken

	// 计算过期时间
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	activeUser.TokenExpiresAt = expiresAt.Unix()

	err = pcsconfig.Config.Save()
	if err != nil {
		fmt.Printf("⚠️  保存配置失败: %v\n", err)
		fmt.Println("💡 请手动保存以下信息:")
		fmt.Printf("AccessToken: %s\n", tokenResp.AccessToken)
		fmt.Printf("RefreshToken: %s\n", tokenResp.RefreshToken)
	} else {
		fmt.Println("✅ 配置已保存")
	}

	return nil
}

// 检查并刷新令牌
func checkAndRefreshToken() error {
	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser.AccessToken == "" {
		return fmt.Errorf("未设置AccessToken，请先登录")
	}

	// 检查是否即将过期（提前5分钟刷新）
	if activeUser.TokenExpiresAt > 0 {
		expiresAt := time.Unix(activeUser.TokenExpiresAt, 0)
		if time.Until(expiresAt) < 5*time.Minute {
			fmt.Println("🔄 AccessToken即将过期，正在刷新...")

			if activeUser.RefreshToken == "" {
				return fmt.Errorf("RefreshToken为空，请重新登录")
			}

			tokenResp, err := refreshAccessToken(DefaultAppKey, DefaultSecretKey, activeUser.RefreshToken)
			if err != nil {
				return fmt.Errorf("刷新令牌失败: %v", err)
			}

			activeUser.AccessToken = tokenResp.AccessToken
			if tokenResp.RefreshToken != "" {
				activeUser.RefreshToken = tokenResp.RefreshToken
			}

			expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			activeUser.TokenExpiresAt = expiresAt.Unix()

			err = pcsconfig.Config.Save()
			if err != nil {
				fmt.Printf("⚠️  保存配置失败: %v\n", err)
			} else {
				fmt.Println("✅ 令牌刷新成功")
			}
		}
	}

	return nil
}
