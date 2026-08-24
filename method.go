
// VerifyMetadataAll 对所有同步歌单执行元数据补全
func (e *Engine) VerifyMetadataAll(systemUserID int) (string, error) {
	enabledPlaylistIDs, err := db.GetEnabledSyncPlaylists(systemUserID)
	if err != nil {
		return "", fmt.Errorf("获取同步歌单列表失败: %w", err)
	}
	if len(enabledPlaylistIDs) == 0 {
		hasSync, _ := db.SyncPlaylistsExist(systemUserID)
		if hasSync {
			return "", fmt.Errorf("没有启用的同步歌单")
		}
		user, err := db.GetCurrentUserForSystem(systemUserID)
		if err != nil || user == nil {
			return "", fmt.Errorf("未登录")
		}
		body, err := e.netease.GetUserPlaylists(user.UserID, user.Cookie)
		if err != nil {
			return "", fmt.Errorf("获取歌单列表失败: %w", err)
		}
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		if playlistData, ok := result["playlist"].([]interface{}); ok {
			for _, p := range playlistData {
				playlist := p.(map[string]interface{})
				if id, ok := playlist["id"].(float64); ok {
					enabledPlaylistIDs = append(enabledPlaylistIDs, int(id))
				}
			}
		}
	}
	if len(enabledPlaylistIDs) == 0 {
		return "", fmt.Errorf("没有找到需要补全的歌单")
	}
	totalTask := e.taskService.CreateTask(service.TaskTypeDataComplete,
		fmt.Sprintf("批量验证补全 %d 个歌单元数据", len(enabledPlaylistIDs)), "")
	e.taskService.SetTaskStatus(totalTask.ID, service.TaskStatusRunning)
	totalTask.Total = len(enabledPlaylistIDs)
	completedCount := 0
	for _, playlistID := range enabledPlaylistIDs {
		taskID, err := e.VerifyMetadata(playlistID)
		if err != nil {
			fmt.Printf("[VerifyMetadataAll] playlist %d failed: %v\n", playlistID, err)
			continue
		}
		fmt.Printf("[VerifyMetadataAll] playlist %d -> task %s\n", playlistID, taskID)
		completedCount++
	}
	e.taskService.UpdateTaskProgress(totalTask.ID, completedCount, len(enabledPlaylistIDs))
	if completedCount == 0 {
		e.taskService.SetTaskStatus(totalTask.ID, service.TaskStatusFailed)
		return "", fmt.Errorf("所有歌单元数据已补全或无可用歌单")
	}
	return totalTask.ID, nil
}
