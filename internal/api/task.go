package api

import (
	"endfield-music/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(ts *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: ts}
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	tasks := h.taskService.GetAllTasks()
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *TaskHandler) GetTaskProgress(c *gin.Context) {
	taskID := c.Param("id")
	progress := h.taskService.GetTaskProgress(taskID)
	c.JSON(http.StatusOK, gin.H{"progress": progress})
}

func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if h.taskService.CancelTask(taskID) {
		c.JSON(http.StatusOK, gin.H{"message": "任务已终止"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务无法终止或不存在"})
	}
}
