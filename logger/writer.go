package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// FileWriter 日志文件写入器
type FileWriter struct {
	logDir       string        // 日志目录
	currentFile  *os.File      // 当前日志文件
	writer       *bufio.Writer // 缓冲写入器
	multiWriter  io.Writer     // 多重写入器（控制台+文件）
	mutex        sync.RWMutex  // 读写锁
	lastRotation time.Time     // 上次轮转时间
	closed       bool          // 是否已关闭
}

// 全局文件写入器实例
var fileWriter *FileWriter
var writerOnce sync.Once

// InitFileLogging 初始化文件日志功能
func InitFileLogging() error {
	var err error
	writerOnce.Do(func() {
		// 获取工作目录
		workDir, dirErr := os.Getwd()
		if dirErr != nil {
			err = fmt.Errorf("获取工作目录失败: %v", dirErr)
			return
		}

		// 创建日志目录
		logDir := filepath.Join(workDir, "logs")
		if mkdirErr := os.MkdirAll(logDir, 0755); mkdirErr != nil {
			err = fmt.Errorf("创建日志目录失败: %v", mkdirErr)
			return
		}

		// 创建文件写入器
		fileWriter = &FileWriter{
			logDir: logDir,
		}

		// 初始化日志文件
		if initErr := fileWriter.initLogFile(); initErr != nil {
			err = fmt.Errorf("初始化日志文件失败: %v", initErr)
			return
		}

		// 设置信号处理
		fileWriter.setupSignalHandler()

		// 启动文件轮转检查
		go fileWriter.rotationChecker()
	})

	if err != nil {
		return err
	}

	// 更新默认日志器的输出
	if fileWriter != nil {
		defaultLogger.output = fileWriter
	}

	return nil
}

// Write 实现 io.Writer 接口（保留兼容性）
func (fw *FileWriter) Write(p []byte) (n int, err error) {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	if fw.closed {
		return 0, fmt.Errorf("文件写入器已关闭")
	}

	// 检查是否需要轮转
	if fw.needsRotation() {
		if rotateErr := fw.rotateFile(); rotateErr != nil {
			// 轮转失败，仍然写入当前文件
			fmt.Fprintf(os.Stderr, "日志文件轮转失败: %v\n", rotateErr)
		}
	}

	// 写入多重写入器
	if fw.multiWriter != nil {
		n, err = fw.multiWriter.Write(p)
		// 立即刷新缓冲区以确保日志及时写入文件
		if fw.writer != nil {
			if flushErr := fw.writer.Flush(); flushErr != nil {
				fmt.Fprintf(os.Stderr, "刷新日志缓冲区失败: %v\n", flushErr)
			}
		}
		return n, err
	}

	return 0, fmt.Errorf("多重写入器未初始化")
}

// WriteLog 自定义日志写入方法，支持不同格式的输出
func (fw *FileWriter) WriteLog(level LogLevel, timestamp, location, message string, keyvals ...interface{}) {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	if fw.closed {
		return
	}

	// 检查是否需要轮转
	if fw.needsRotation() {
		if rotateErr := fw.rotateFile(); rotateErr != nil {
			fmt.Fprintf(os.Stderr, "日志文件轮转失败: %v\n", rotateErr)
		}
	}

	// 格式化控制台日志（带颜色）
	consoleLog := formatLogForConsole(level, timestamp, location, message, keyvals...)
	// 格式化文件日志（不带颜色）
	fileLog := formatLogForFile(level, timestamp, location, message, keyvals...)

	// 写入控制台
	fmt.Fprintln(os.Stdout, consoleLog)

	// 写入文件
	if fw.writer != nil {
		fmt.Fprintln(fw.writer, fileLog)
		// 立即刷新缓冲区
		if flushErr := fw.writer.Flush(); flushErr != nil {
			fmt.Fprintf(os.Stderr, "刷新日志缓冲区失败: %v\n", flushErr)
		}
	}
}

// initLogFile 初始化日志文件
func (fw *FileWriter) initLogFile() error {
	fileName := fw.generateFileName(time.Now())
	filePath := filepath.Join(fw.logDir, fileName)

	// 打开或创建日志文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	fw.currentFile = file
	fw.writer = bufio.NewWriter(file)
	fw.lastRotation = time.Now()

	return nil
}

// generateFileName 生成日志文件名
func (fw *FileWriter) generateFileName(t time.Time) string {
	dateStr := t.Format("2006-01-02")
	var period string

	// 判断是上午还是下午（12小时制）
	if t.Hour() < 12 {
		period = "AM"
	} else {
		period = "PM"
	}

	return fmt.Sprintf("app_%s_%s.log", dateStr, period)
}

// needsRotation 检查是否需要轮转文件
func (fw *FileWriter) needsRotation() bool {
	now := time.Now()

	// 检查是否跨越了12小时边界
	lastHour := fw.lastRotation.Hour()
	currentHour := now.Hour()

	// 如果跨越了0点或12点，需要轮转
	if (lastHour < 12 && currentHour >= 12) ||
		(fw.lastRotation.Day() != now.Day()) {
		return true
	}

	return false
}

// rotateFile 轮转日志文件
func (fw *FileWriter) rotateFile() error {
	// 刷新并关闭当前文件
	if fw.writer != nil {
		if err := fw.writer.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "刷新缓冲区失败: %v\n", err)
		}
	}

	if fw.currentFile != nil {
		if err := fw.currentFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭当前日志文件失败: %v\n", err)
		}
	}

	// 创建新的日志文件
	return fw.initLogFile()
}

// rotationChecker 定期检查文件轮转
func (fw *FileWriter) rotationChecker() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for range ticker.C {
		fw.mutex.Lock()
		if !fw.closed && fw.needsRotation() {
			if err := fw.rotateFile(); err != nil {
				fmt.Fprintf(os.Stderr, "定时轮转失败: %v\n", err)
			}
		}
		fw.mutex.Unlock()
	}
}

// setupSignalHandler 设置信号处理器
func (fw *FileWriter) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n接收到退出信号，正在保存日志...")
		fw.Close()
		os.Exit(0)
	}()
}

// Flush 刷新缓冲区
func (fw *FileWriter) Flush() error {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	if fw.writer != nil {
		return fw.writer.Flush()
	}
	return nil
}

// Close 关闭文件写入器
func (fw *FileWriter) Close() error {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	if fw.closed {
		return nil
	}

	fw.closed = true

	// 刷新缓冲区
	if fw.writer != nil {
		if err := fw.writer.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "最终刷新失败: %v\n", err)
		}
	}

	// 关闭文件
	if fw.currentFile != nil {
		if err := fw.currentFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭日志文件失败: %v\n", err)
			return err
		}
	}

	fmt.Println("日志文件已安全关闭")
	return nil
}

// GetFileWriter 获取文件写入器实例
func GetFileWriter() *FileWriter {
	return fileWriter
}

// FlushLogs 手动刷新日志
func FlushLogs() error {
	if fileWriter != nil {
		return fileWriter.Flush()
	}
	return nil
}
