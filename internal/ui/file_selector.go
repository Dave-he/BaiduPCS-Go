package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileSearchItem 文件搜索结果项
type FileSearchItem struct {
	FsId         uint64 `json:"fs_id"`
	Path         string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Size         uint64 `json:"size"`
	IsDir        int    `json:"isdir"`
	Category     int    `json:"category"`
	Md5          string `json:"md5"`
	ServerMtime  int64  `json:"server_mtime"`
}

// FileSelector 交互式文件选择器
type FileSelector struct {
	files []FileSearchItem
}

// NewFileSelector 创建文件选择器
func NewFileSelector(files []FileSearchItem) *FileSelector {
	return &FileSelector{files: files}
}

// SelectFile 显示文件列表并获取用户选择
func (fs *FileSelector) SelectFile() (*FileSearchItem, error) {
	if len(fs.files) == 0 {
		return nil, fmt.Errorf("没有找到匹配的文件")
	}

	if len(fs.files) == 1 {
		fmt.Printf("✅ 找到唯一匹配文件: %s\n", fs.files[0].ServerFilename)
		return &fs.files[0], nil
	}

	// 显示文件列表
	fmt.Printf("\n🔍 找到 %d 个匹配的文件:\n", len(fs.files))
	fmt.Println(strings.Repeat("=", 80))
	
	for i, file := range fs.files {
		sizeStr := FormatFileSize(file.Size)
		fmt.Printf("  %d. 📄 %s\n", i+1, file.ServerFilename)
		fmt.Printf("     📏 大小: %s\n", sizeStr)
		fmt.Printf("     📂 路径: %s\n", file.Path)
		if file.Md5 != "" {
			fmt.Printf("     🔐 MD5: %s\n", file.Md5)
		}
		fmt.Println(strings.Repeat("-", 80))
	}

	// 获取用户选择
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("\n请选择要下载的文件 (1-%d，输入 0 取消): ", len(fs.files))
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取输入失败: %v", err)
		}

		input = strings.TrimSpace(input)
		if input == "0" {
			return nil, fmt.Errorf("用户取消操作")
		}

		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("❌ 请输入有效的数字")
			continue
		}

		if choice < 1 || choice > len(fs.files) {
			fmt.Printf("❌ 请输入 1-%d 之间的数字\n", len(fs.files))
			continue
		}

		selectedFile := &fs.files[choice-1]
		fmt.Printf("✅ 已选择: %s\n", selectedFile.ServerFilename)
		return selectedFile, nil
	}
}

// GenerateOutputPath 智能输出路径生成
func GenerateOutputPath(filename, userOutput string) string {
	if userOutput == "" {
		// 如果没有指定输出路径，使用文件自身的名字
		return filename
	}
	
	// 检查用户指定的路径是否存在且为目录
	if stat, err := os.Stat(userOutput); err == nil && stat.IsDir() {
		// 如果是目录，在该目录下使用原文件名
		return filepath.Join(userOutput, filename)
	}
	
	// 检查用户指定的路径是否以路径分隔符结尾，表示这是一个目录
	if strings.HasSuffix(userOutput, "/") || strings.HasSuffix(userOutput, "\\") {
		// 确保目录存在
		os.MkdirAll(userOutput, 0755)
		return filepath.Join(userOutput, filename)
	}
	
	// 检查父目录是否存在，如果不存在则创建
	parentDir := filepath.Dir(userOutput)
	if parentDir != "." && parentDir != "" {
		os.MkdirAll(parentDir, 0755)
	}
	
	// 否则将用户指定的路径作为完整的文件路径
	return userOutput
}

// ConfirmDownload 确认下载操作
func ConfirmDownload(filename, outputPath string, fileSize uint64) bool {
	fmt.Printf("\n📋 下载确认:\n")
	fmt.Printf("  文件名: %s\n", filename)
	fmt.Printf("  大小: %s\n", FormatFileSize(fileSize))
	fmt.Printf("  保存到: %s\n", outputPath)
	
	// 检查文件是否已存在
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("⚠️  目标文件已存在，将被覆盖\n")
	}
	
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n确认下载? (y/n): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		
		input = strings.ToLower(strings.TrimSpace(input))
		switch input {
		case "y", "yes", "是":
			return true
		case "n", "no", "否":
			return false
		default:
			fmt.Println("请输入 y 或 n")
		}
	}
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

// GetFileCategory 文件类型检测
func GetFileCategory(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return "🖼️  图片"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv":
		return "🎬 视频"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "🎵 音频"
	case ".pdf":
		return "📄 PDF"
	case ".doc", ".docx":
		return "📝 Word文档"
	case ".xls", ".xlsx":
		return "📊 Excel表格"
	case ".ppt", ".pptx":
		return "📽️  PowerPoint"
	case ".txt":
		return "📄 文本文件"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "📦 压缩包"
	default:
		return "📄 文件"
	}
}