package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// 错误处理器
type ErrorHandler struct {
	logger *log.Logger
}

// 创建错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		logger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
	}
}

// 处理错误并记录
func (eh *ErrorHandler) HandleError(err error, context string) {
	if err != nil {
		eh.logger.Printf("%s: %v", context, err)
	}
}

// 处理致命错误
func (eh *ErrorHandler) HandleFatalError(err error, context string) {
	if err != nil {
		eh.logger.Fatalf("%s: %v", context, err)
	}
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	startTime time.Time
	name      string
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(name string) *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime: time.Now(),
		name:      name,
	}
}

// End 结束监控并输出结果
func (pm *PerformanceMonitor) End() {
	duration := time.Since(pm.startTime)
	fmt.Printf("⏱️  %s 耗时: %v\n", pm.name, duration)
}

// PrintMemoryUsage 内存使用监控
func PrintMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	fmt.Printf("💾 内存使用情况:\n")
	fmt.Printf("  分配内存: %d KB\n", bToKb(m.Alloc))
	fmt.Printf("  总分配: %d KB\n", bToKb(m.TotalAlloc))
	fmt.Printf("  系统内存: %d KB\n", bToKb(m.Sys))
	fmt.Printf("  GC次数: %d\n", m.NumGC)
}

// bToKb 字节转KB
func bToKb(b uint64) uint64 {
	return b / 1024
}

// RetryOperation 重试机制
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

// ValidateFileSize 文件大小验证
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

// SafeCreateFile 安全的文件创建
func SafeCreateFile(filePath string) (*os.File, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}
	
	// 如果文件已存在，创建备份
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filePath + ".backup." + time.Now().Format("20060102150405")
		if err := os.Rename(filePath, backupPath); err != nil {
			fmt.Printf("⚠️  无法创建备份文件: %v\n", err)
		} else {
			fmt.Printf("📋 已创建备份文件: %s\n", backupPath)
		}
	}
	
	return os.Create(filePath)
}

// CleanupTempFiles 清理临时文件
func CleanupTempFiles(patterns []string) {
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		
		for _, match := range matches {
			if err := os.Remove(match); err == nil {
				fmt.Printf("🗑️  清理临时文件: %s\n", match)
			}
		}
	}
}