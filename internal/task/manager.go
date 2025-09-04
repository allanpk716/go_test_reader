package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/allanpk716/go_test_reader/internal/parser"
)

// Status 任务状态
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// Task 分析任务
type Task struct {
	ID        string
	FilePath  string
	Status    Status
	Result    *parser.TestResult
	Error     error
	CreatedAt time.Time
	UpdatedAt time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// Manager 任务管理器
type Manager struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

// NewManager 创建新的任务管理器
func NewManager() *Manager {
	return &Manager{
		tasks: make(map[string]*Task),
	}
}

// CreateTask 创建新任务
func (m *Manager) CreateTask(id, filePath string) *Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	task := &Task{
		ID:        id,
		FilePath:  filePath,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	m.tasks[id] = task
	return task
}

// GetTask 获取任务
func (m *Manager) GetTask(id string) *Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.tasks[id]
}

// TerminateTask 终止任务
func (m *Manager) TerminateTask(id string) error {
	m.mu.RLock()
	task, exists := m.tasks[id]
	m.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}
	
	task.mu.Lock()
	defer task.mu.Unlock()
	
	if task.Status == StatusCompleted || task.Status == StatusFailed || task.Status == StatusCanceled {
		return fmt.Errorf("task %s is already finished with status: %s", id, task.Status)
	}
	
	task.cancel()
	task.Status = StatusCanceled
	task.UpdatedAt = time.Now()
	
	return nil
}

// ListTasks 列出所有任务
func (m *Manager) ListTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tasks := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	
	return tasks
}

// CleanupOldTasks 清理旧任务（可选的后台清理功能）
func (m *Manager) CleanupOldTasks(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	for id, task := range m.tasks {
		if now.Sub(task.CreatedAt) > maxAge {
			task.mu.Lock()
			if task.cancel != nil {
				task.cancel()
			}
			task.mu.Unlock()
			delete(m.tasks, id)
		}
	}
}

// SetRunning 设置任务为运行状态
func (t *Task) SetRunning() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.Status = StatusRunning
	t.UpdatedAt = time.Now()
}

// SetResult 设置任务结果
func (t *Task) SetResult(result *parser.TestResult) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.Result = result
	t.Status = StatusCompleted
	t.UpdatedAt = time.Now()
}

// SetError 设置任务错误
func (t *Task) SetError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.Error = err
	t.Status = StatusFailed
	t.UpdatedAt = time.Now()
}

// GetStatus 获取任务状态信息
func (t *Task) GetStatus() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	status := map[string]interface{}{
		"task_id":    t.ID,
		"status":     string(t.Status),
		"file_path":  t.FilePath,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
	
	if t.Error != nil {
		status["error"] = t.Error.Error()
	}
	
	if t.Result != nil {
		status["result"] = map[string]interface{}{
			"total_tests":   t.Result.TotalTests,
			"passed_tests":  t.Result.PassedTests,
			"failed_tests":  t.Result.FailedTests,
			"skipped_tests": t.Result.SkippedTests,
			"failed_test_names": t.Result.FailedTestNames,
			"passed_test_names": t.Result.PassedTestNames,
		}
	}
	
	return status
}

// GetTestDetails 获取特定测试的详细信息
func (t *Task) GetTestDetails(testName string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if t.Result == nil {
		return nil
	}
	
	if details, exists := t.Result.TestDetails[testName]; exists {
		return map[string]interface{}{
			"test_name": testName,
			"status":    details.Status,
			"output":    details.Output,
			"error":     details.Error,
			"elapsed":   details.Elapsed,
		}
	}
	
	return nil
}

// IsCanceled 检查任务是否被取消
func (t *Task) IsCanceled() bool {
	select {
	case <-t.ctx.Done():
		return true
	default:
		return false
	}
}