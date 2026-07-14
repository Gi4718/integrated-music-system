package service

import (
	"sort"
	"sync"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskType string

const (
	TaskTypeDownload      TaskType = "download"
	TaskTypeDataComplete  TaskType = "data_complete"
	TaskTypeSync          TaskType = "sync"
	TaskTypeScan          TaskType = "scan"
)

type Task struct {
	ID            string     `json:"id"`
	Type          TaskType   `json:"type"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        TaskStatus `json:"status"`
	Progress      int        `json:"progress"` // 0-100
	Total         int        `json:"total"`
	Current       int        `json:"current"`
	CurrentFile   string     `json:"current_file"`
	CurrentBytes  int64      `json:"current_bytes"`
	TotalBytes    int64      `json:"total_bytes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Error         string     `json:"error,omitempty"`
}

type TaskService struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewTaskService() *TaskService {
	return &TaskService{
		tasks: make(map[string]*Task),
	}
}

func (s *TaskService) CreateTask(taskType TaskType, title, description string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateTaskID()
	task := &Task{
		ID:          id,
		Type:        taskType,
		Title:       title,
		Description: description,
		Status:      TaskStatusPending,
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.tasks[id] = task
	return task
}

func (s *TaskService) UpdateTaskProgress(id string, current, total int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.Current = current
		task.Total = total
		if total > 0 {
			task.Progress = (current * 100) / total
		}
		task.Status = TaskStatusRunning
		task.UpdatedAt = time.Now()
	}
}

func (s *TaskService) CompleteTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.Status = TaskStatusCompleted
		task.Progress = 100
		task.UpdatedAt = time.Now()
	}
}

func (s *TaskService) FailTask(id string, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.Status = TaskStatusFailed
		task.Error = errMsg
		task.UpdatedAt = time.Now()
	}
}

func (s *TaskService) GetAllTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		result = append(result, task)
	}
	// 按创建时间升序排序，保证任务日志顺序稳定
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

func (s *TaskService) GetTaskProgress(id string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, ok := s.tasks[id]; ok {
		return task
	}
	return nil
}

func (s *TaskService) UpdateTaskCurrentFile(id string, filename string, downloaded, total int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.CurrentFile = filename
		task.CurrentBytes = downloaded
		task.TotalBytes = total
		task.UpdatedAt = time.Now()
	}
}

func (s *TaskService) SetTaskStatus(id string, status TaskStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.Status = status
		task.UpdatedAt = time.Now()
	}
}

func (s *TaskService) GetTask(id string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, ok := s.tasks[id]; ok {
		return task
	}
	return nil
}

func (s *TaskService) CancelTask(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		if task.Status == TaskStatusPending || task.Status == TaskStatusRunning {
			task.Status = TaskStatusCancelled
			task.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

func generateTaskID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}
