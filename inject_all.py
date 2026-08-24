#!/usr/bin/env python3
"""Inject VerifyMetadataAll into all feature branch files"""
import sys

def inject_engine(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'VerifyMetadataAll' in content:
        print(f"  engine.go: already has VerifyMetadataAll")
        return

    method = '''
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

'''
    marker = "func (e *Engine) failTask("
    if marker in content:
        content = content.replace(marker, method + marker)
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"  engine.go: OK - VerifyMetadataAll added")
    else:
        print(f"  engine.go: ERROR - marker not found")

def inject_download(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'verifyMetadataAll' in content:
        print(f"  download.go: already has verifyMetadataAll")
        return

    handler = '''
func verifyMetadataAll(c *gin.Context) {
\tengine := getDownloadEngine()
\tif engine == nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "下载引擎未初始化"})
\t\treturn
\t}

\tsystemUserID := getSystemUserID(c)
\ttaskID, err := engine.VerifyMetadataAll(systemUserID)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
\t\treturn
\t}

\tc.JSON(http.StatusOK, gin.H{
\t\t"message": "批量验证补全任务已创建",
\t\t"task_id": taskID,
\t})
}
'''
    # Add after the last closing brace of verifyMetadata
    marker = 'c.JSON(http.StatusOK, gin.H{\n\t\t"message": "验证补全任务已创建",'
    if marker in content:
        # Find the end of verifyMetadata function
        idx = content.index(marker)
        # Find the closing })
        rest = content[idx:]
        end_idx = rest.index('})') + len('})')
        insert_pos = idx + end_idx
        content = content[:insert_pos] + '\n' + handler + content[insert_pos:]
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"  download.go: OK - verifyMetadataAll added")
    else:
        print(f"  download.go: ERROR - marker not found")

def inject_router(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'verify-metadata-all' in content:
        print(f"  router.go: already has verify-metadata-all")
        return

    line = '\t\t\t\tdownload.POST("/verify-metadata-all", verifyMetadataAll)'
    marker = 'download.POST("/verify-metadata", verifyMetadata)'
    if marker in content:
        content = content.replace(marker, marker + '\n' + line)
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"  router.go: OK - route added")
    else:
        print(f"  router.go: ERROR - marker not found")

def inject_api_index(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'verifyMetadataAll' in content:
        print(f"  api/index.ts: already has verifyMetadataAll")
        return

    marker = "verifyMetadata(playlistId: number) {"
    replacement = """verifyMetadata(playlistId: number) {
      return api.post('/download/verify-metadata', { playlist_id: playlistId })
    },
    verifyMetadataAll() {
      return api.post('/download/verify-metadata-all')
    }"""
    # Find and replace the verifyMetadata method including its body
    old = """verifyMetadata(playlistId: number) {
      return api.post('/download/verify-metadata', { playlist_id: playlistId })
    }"""
    if old in content:
        content = content.replace(old, replacement)
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"  api/index.ts: OK - verifyMetadataAll added")
    else:
        print(f"  api/index.ts: ERROR - marker not found")

if __name__ == '__main__':
    base = sys.argv[1]
    print("Injecting into engine.go...")
    inject_engine(f'{base}/internal/download/engine.go')
    print("Injecting into download.go...")
    inject_download(f'{base}/internal/api/download.go')
    print("Injecting into router.go...")
    inject_router(f'{base}/internal/api/router.go')
    print("Injecting into api/index.ts...")
    inject_api_index(f'{base}/web/src/api/index.ts')
    print("Done!")
