package task_manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
)

// TaskStatus 任务状态枚举
type TaskStatus int

const (
	TaskStatusIdle    TaskStatus = iota // 空闲状态
	TaskStatusRunning                   // 任务运行中
)

// TaskInfo 任务信息
type TaskInfo struct {
	ID        string    // 任务ID
	StartTime time.Time // 任务开始时间
	Texts     []string  // 任务期间收集的文本
}

// TaskManager 任务管理器
type TaskManager struct {
	mutex       sync.RWMutex // 读写锁保护并发访问
	status      TaskStatus   // 当前任务状态
	currentTask *TaskInfo    // 当前任务信息
	textWindow  []string     // 文本窗口
	taskCounter int          // 任务计数器，用于生成任务ID
}

var (
	instance *TaskManager
	once     sync.Once
)

// GetInstance 获取任务管理器单例实例
func GetInstance() *TaskManager {
	once.Do(func() {
		instance = &TaskManager{
			status:      TaskStatusIdle,
			currentTask: nil,
			textWindow:  make([]string, 0),
			taskCounter: 0,
		}
		logger.Info("任务管理器初始化完成")
	})
	return instance
}

// AddText 添加文本到窗口
// 当窗口为空时，第一个文本的添加会自动开始新任务
func (tm *TaskManager) AddText(text string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if text == "" {
		return fmt.Errorf("文本内容不能为空")
	}

	// 检查是否需要开始新任务
	if tm.status == TaskStatusIdle && len(tm.textWindow) == 0 {
		tm.startNewTask()
	}

	// 添加文本到窗口
	tm.textWindow = append(tm.textWindow, text)

	// 如果有当前任务，也添加到任务记录中
	if tm.currentTask != nil {
		tm.currentTask.Texts = append(tm.currentTask.Texts, text)
	}

	logger.Info(fmt.Sprintf("添加文本到窗口: %s (窗口大小: %d)", text, len(tm.textWindow)))
	return nil
}

// startNewTask 开始新任务（内部方法，调用前需要加锁）
func (tm *TaskManager) startNewTask() {
	tm.taskCounter++
	taskID := fmt.Sprintf("task_%d_%d", tm.taskCounter, time.Now().Unix())

	tm.currentTask = &TaskInfo{
		ID:        taskID,
		StartTime: time.Now(),
		Texts:     make([]string, 0),
	}
	tm.status = TaskStatusRunning

	logger.Info(fmt.Sprintf("开始新任务: %s", taskID))
}

// CompleteTask 完成当前任务并返回所有文本信息
// 返回窗口中的所有文本，并清空窗口准备下一个任务
func (tm *TaskManager) CompleteTask() []string {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.status == TaskStatusIdle {
		logger.Warn("没有正在运行的任务")
		return []string{}
	}

	// 获取窗口中的所有文本
	texts := make([]string, len(tm.textWindow))
	copy(texts, tm.textWindow)

	// 记录任务完成信息
	if tm.currentTask != nil {
		duration := time.Since(tm.currentTask.StartTime)
		logger.Info(fmt.Sprintf("任务完成: %s, 持续时间: %v, 收集文本数量: %d",
			tm.currentTask.ID, duration, len(texts)))
	}

	// 清空窗口和重置状态
	tm.textWindow = tm.textWindow[:0] // 清空slice但保留容量
	tm.status = TaskStatusIdle
	tm.currentTask = nil

	logger.Info("任务窗口已清空，准备接收下一个任务")
	return texts
}

// IsTaskRunning 检查是否有任务在运行
func (tm *TaskManager) IsTaskRunning() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return tm.status == TaskStatusRunning
}

// GetCurrentTexts 获取当前窗口中的文本（不清空窗口）
func (tm *TaskManager) GetCurrentTexts() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// 返回文本的副本，避免外部修改
	texts := make([]string, len(tm.textWindow))
	copy(texts, tm.textWindow)

	return texts
}

// GetTaskInfo 获取当前任务信息
func (tm *TaskManager) GetTaskInfo() *TaskInfo {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if tm.currentTask == nil {
		return nil
	}

	// 返回任务信息的副本
	taskInfo := &TaskInfo{
		ID:        tm.currentTask.ID,
		StartTime: tm.currentTask.StartTime,
		Texts:     make([]string, len(tm.currentTask.Texts)),
	}
	copy(taskInfo.Texts, tm.currentTask.Texts)

	return taskInfo
}

// GetWindowSize 获取当前窗口大小
func (tm *TaskManager) GetWindowSize() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return len(tm.textWindow)
}

// ForceCompleteTask 强制完成当前任务（即使窗口为空）
func (tm *TaskManager) ForceCompleteTask() []string {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	texts := make([]string, len(tm.textWindow))
	copy(texts, tm.textWindow)

	if tm.currentTask != nil {
		duration := time.Since(tm.currentTask.StartTime)
		logger.Info(fmt.Sprintf("强制完成任务: %s, 持续时间: %v, 收集文本数量: %d",
			tm.currentTask.ID, duration, len(texts)))
	}

	// 清空窗口和重置状态
	tm.textWindow = tm.textWindow[:0]
	tm.status = TaskStatusIdle
	tm.currentTask = nil

	return texts
}

// ClearWindow 清空窗口但不完成任务（紧急情况使用）
func (tm *TaskManager) ClearWindow() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	oldSize := len(tm.textWindow)
	tm.textWindow = tm.textWindow[:0]

	logger.Warn(fmt.Sprintf("窗口已被强制清空，丢失 %d 条文本", oldSize))
}

// GetStats 获取任务管理器统计信息
func (tm *TaskManager) GetStats() map[string]interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	stats := map[string]interface{}{
		"status":       tm.status,
		"window_size":  len(tm.textWindow),
		"task_counter": tm.taskCounter,
	}

	if tm.currentTask != nil {
		stats["current_task_id"] = tm.currentTask.ID
		stats["current_task_start_time"] = tm.currentTask.StartTime
		stats["current_task_duration"] = time.Since(tm.currentTask.StartTime)
		stats["current_task_texts_count"] = len(tm.currentTask.Texts)
	}

	return stats
}

// 便利函数，直接使用单例实例

// AddText 添加文本到全局任务管理器
func AddText(text string) error {
	return GetInstance().AddText(text)
}

// CompleteTask 完成当前任务
func CompleteTask() []string {
	return GetInstance().CompleteTask()
}

// IsTaskRunning 检查是否有任务在运行
func IsTaskRunning() bool {
	return GetInstance().IsTaskRunning()
}

// GetCurrentTexts 获取当前窗口中的文本
func GetCurrentTexts() []string {
	return GetInstance().GetCurrentTexts()
}

// GetTaskInfo 获取当前任务信息
func GetTaskInfo() *TaskInfo {
	return GetInstance().GetTaskInfo()
}

// GetWindowSize 获取当前窗口大小
func GetWindowSize() int {
	return GetInstance().GetWindowSize()
}

// GetStats 获取统计信息
func GetStats() map[string]interface{} {
	return GetInstance().GetStats()
}
