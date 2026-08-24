#!/usr/bin/env python3
import sys

filepath = sys.argv[1]
with open(filepath, 'r') as f:
    content = f.read()

method = """
// VerifyMetadataAll 对所有同步歌单执行元数据补全
func (e *Engine) VerifyMetadataAll(systemUserID int) (string, error) {
\tenabledPlaylistIDs, err := db.GetEnabledSyncPlaylists(systemUserID)
\tif err != nil {
\t\treturn "", fmt.Errorf("获取同步歌单列表失败: %w", err)
\t}
\tif len(enabledPlaylistIDs) == 0 {
\t\thasSync, _ := db.SyncPlaylistsExist(systemUserID)
\t\tif hasSync {
\t\t\treturn "", fmt.Errorf("没有启用的同步歌单")
\t\t}
\t\tuser, err := db.GetCurrentUserForSystem(systemUserID)
\t\tif err != nil || user == nil {
\t\t\treturn "", fmt.Errorf("未登录")
\t\t}
\t\tbody, err := e.netease.GetUserPlaylists(user.UserID, user.Cookie)
\t\tif err != nil {
\t\t\treturn "", fmt.Errorf("获取歌单列表失败: %w", err)
\t\t}
\t\tvar result map[string]interface{}
\t\tjson.Unmarshal(body, &result)
\t\tif playlistData, ok := result["playlist"].([]interface{}); ok {
\t\t\tfor _, p := range playlistData {
\t\t\t\tplaylist := p.(map[string]interface{})
\t\t\t\tif id, ok := playlist["id"].(float64); ok {
\t\t\t\t\tenabledPlaylistIDs = append(enabledPlaylistIDs, int(id))
\t\t\t\t}
\t\t\t}
\t\t}
\t}
\tif len(enabledPlaylistIDs) == 0 {
\t\treturn "", fmt.Errorf("没有找到需要补全的歌单")
\t}
\ttotalTask := e.taskService.CreateTask(service.TaskTypeDataComplete,
\t\tfmt.Sprintf("批量验证补全 %d 个歌单元数据", len(enabledPlaylistIDs)), "")
\te.taskService.SetTaskStatus(totalTask.ID, service.TaskStatusRunning)
\ttotalTask.Total = len(enabledPlaylistIDs)
\tcompletedCount := 0
\tfor _, playlistID := range enabledPlaylistIDs {
\t\ttaskID, err := e.VerifyMetadata(playlistID)
\t\tif err != nil {
\t\t\tfmt.Printf("[VerifyMetadataAll] playlist %d failed: %v\\n", playlistID, err)
\t\t\tcontinue
\t\t}
\t\tfmt.Printf("[VerifyMetadataAll] playlist %d -> task %s\\n", playlistID, taskID)
\t\tcompletedCount++
\t}
\te.taskService.UpdateTaskProgress(totalTask.ID, completedCount, len(enabledPlaylistIDs))
\tif completedCount == 0 {
\t\te.taskService.SetTaskStatus(totalTask.ID, service.TaskStatusFailed)
\t\treturn "", fmt.Errorf("所有歌单元数据已补全或无可用歌单")
\t}
\treturn totalTask.ID, nil
}

"""

marker = "func (e *Engine) failTask("
if marker in content:
    content = content.replace(marker, method + marker)
    with open(filepath, 'w') as f:
        f.write(content)
    print("OK: VerifyMetadataAll added")
else:
    print("ERROR: marker not found")
