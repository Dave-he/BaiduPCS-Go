package ui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// 进度条结构体
type ProgressBar struct {
	total       int64
	current     int64
	width       int
	startTime   time.Time
	lastUpdate  time.Time
	mutex       sync.Mutex
	description string
}

// 创建新的进度条
func NewProgressBar(total int64, description string) *ProgressBar {
	return &ProgressBar{
		total:       total,
		current:     0,
		width:       50,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
		description: description,
	}
}

// 更新进度
func (pb *ProgressBar) Update(current int64) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()
	
	pb.current = current
	now := time.Now()
	
	// 限制更新频率，避免过于频繁的输出
	if now.Sub(pb.lastUpdate) < 100*time.Millisecond && current < pb.total {
		return
	}
	pb.lastUpdate = now
	
	pb.render()
}

// 完成进度条
func (pb *ProgressBar) Finish() {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()
	
	pb.current = pb.total
	pb.render()
	fmt.Println() // 换行
}

// 渲染进度条
func (pb *ProgressBar) render() {
	if pb.total <= 0 {
		return
	}
	
	percentage := float64(pb.current) / float64(pb.total) * 100
	filled := int(float64(pb.width) * float64(pb.current) / float64(pb.total))
	
	// 构建进度条
	bar := "["
	for i := 0; i < pb.width; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"
	
	// 计算速度和剩余时间
	elapsed := time.Since(pb.startTime)
	var speed float64
	var eta string
	
	if elapsed.Seconds() > 0 {
		speed = float64(pb.current) / elapsed.Seconds()
		if speed > 0 && pb.current < pb.total {
			remaining := float64(pb.total-pb.current) / speed
			eta = fmt.Sprintf(" ETA: %s", time.Duration(remaining*float64(time.Second)).Round(time.Second))
		} else {
			eta = ""
		}
	}
	
	// 格式化大小和速度
	currentStr := FormatFileSize(uint64(pb.current))
	totalStr := FormatFileSize(uint64(pb.total))
	speedStr := FormatFileSize(uint64(speed)) + "/s"
	
	// 输出进度条
	fmt.Printf("\r%s %s %.1f%% (%s/%s) %s%s",
		pb.description, bar, percentage, currentStr, totalStr, speedStr, eta)
}

// 进度读取器 - 包装io.Reader以跟踪读取进度
type ProgressReader struct {
	reader io.Reader
	pb     *ProgressBar
	total  int64
	read   int64
}

// 创建进度读取器
func NewProgressReader(reader io.Reader, total int64, description string) *ProgressReader {
	return &ProgressReader{
		reader: reader,
		pb:     NewProgressBar(total, description),
		total:  total,
		read:   0,
	}
}

// 实现io.Reader接口
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.read += int64(n)
	pr.pb.Update(pr.read)
	
	if err == io.EOF {
		pr.pb.Finish()
	}
	
	return n, err
}

// 进度写入器 - 包装io.Writer以跟踪写入进度
type ProgressWriter struct {
	writer io.Writer
	pb     *ProgressBar
	total  int64
	written int64
}

// 创建进度写入器
func NewProgressWriter(writer io.Writer, total int64, description string) *ProgressWriter {
	return &ProgressWriter{
		writer: writer,
		pb:     NewProgressBar(total, description),
		total:  total,
		written: 0,
	}
}

// 实现io.Writer接口
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.written += int64(n)
	pw.pb.Update(pw.written)
	
	if pw.written >= pw.total {
		pw.pb.Finish()
	}
	
	return n, err
}

// 获取进度条
func (pw *ProgressWriter) GetProgressBar() *ProgressBar {
	return pw.pb
}

// 多文件进度管理器
type MultiProgressManager struct {
	bars   []*ProgressBar
	mutex  sync.Mutex
	active bool
}

// 创建多进度管理器
func NewMultiProgressManager() *MultiProgressManager {
	return &MultiProgressManager{
		bars:   make([]*ProgressBar, 0),
		active: true,
	}
}

// 添加进度条
func (mpm *MultiProgressManager) AddProgressBar(pb *ProgressBar) {
	mpm.mutex.Lock()
	defer mpm.mutex.Unlock()
	mpm.bars = append(mpm.bars, pb)
}

// 停止所有进度条
func (mpm *MultiProgressManager) Stop() {
	mpm.mutex.Lock()
	defer mpm.mutex.Unlock()
	mpm.active = false
}

// 渲染所有进度条
func (mpm *MultiProgressManager) Render() {
	mpm.mutex.Lock()
	defer mpm.mutex.Unlock()
	
	if !mpm.active {
		return
	}
	
	// 清屏并重新绘制所有进度条
	fmt.Print("\033[2J\033[H") // 清屏并移动光标到左上角
	
	for _, bar := range mpm.bars {
		bar.render()
		fmt.Println()
	}
}