# 注册系统用户并测试

$baseUrl = "http://192.168.1.70:33550"

Write-Host "1. 检查系统用户状态" -ForegroundColor Cyan
try {
    $checkResult = Invoke-RestMethod -Uri "$baseUrl/api/system/check" -Method GET
    Write-Host "系统用户状态: $($checkResult.has_user)" -ForegroundColor Yellow
} catch {
    Write-Host "检查失败: $_" -ForegroundColor Red
}

Write-Host "`n2. 注册系统用户" -ForegroundColor Cyan
$registerBody = @{username="89913969@qq.com"; password="Luo122115"} | ConvertTo-Json
try {
    $registerResult = Invoke-RestMethod -Uri "$baseUrl/api/system/register" -Method POST -Body $registerBody -ContentType "application/json"
    Write-Host "注册成功" -ForegroundColor Green
    $registerResult | ConvertTo-Json
} catch {
    Write-Host "注册失败: $_" -ForegroundColor Yellow
}

Write-Host "`n3. 登录系统" -ForegroundColor Cyan
$loginBody = @{username="89913969@qq.com"; password="Luo122115"} | ConvertTo-Json
try {
    $loginResult = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "登录成功" -ForegroundColor Green
    $token = $loginResult.token
    Write-Host "Token: $($token.Substring(0,50))..." -ForegroundColor Green
} catch {
    Write-Host "登录失败: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`n4. 获取当前设置" -ForegroundColor Cyan
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "resume_downloads: $($settings.settings.resume_downloads)"
Write-Host "delete_removed: $($settings.settings.delete_removed)"
Write-Host "sync_mode: $($settings.settings.sync_mode)"

Write-Host "`n5. 测试 resume_downloads 开关" -ForegroundColor Cyan
Write-Host "设置 resume_downloads = false"
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body '{"resume_downloads":"false"}' -ContentType "application/json"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "验证: resume_downloads = $($settings.settings.resume_downloads)" -ForegroundColor Green

Write-Host "`n设置 resume_downloads = true"
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body '{"resume_downloads":"true"}' -ContentType "application/json"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "验证: resume_downloads = $($settings.settings.resume_downloads)" -ForegroundColor Green

Write-Host "`n6. 测试 delete_removed 开关" -ForegroundColor Cyan
Write-Host "设置 delete_removed = true"
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body '{"delete_removed":"true"}' -ContentType "application/json"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "验证: delete_removed = $($settings.settings.delete_removed)" -ForegroundColor Green

Write-Host "`n恢复 delete_removed = false"
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body '{"delete_removed":"false"}' -ContentType "application/json"

Write-Host "`n7. 测试 sync_mode 定时模式" -ForegroundColor Cyan
Write-Host "设置 sync_mode = schedule"
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body '{"sync_mode":"schedule","sync_weekdays":"[1,3,5]","sync_time":"08:30"}' -ContentType "application/json"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "验证: sync_mode = $($settings.settings.sync_mode), sync_weekdays = $($settings.settings.sync_weekdays), sync_time = $($settings.settings.sync_time)" -ForegroundColor Green

Write-Host "`n恢复 sync_mode = interval"
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body '{"sync_mode":"interval"}' -ContentType "application/json"

Write-Host "`n8. 测试下载功能" -ForegroundColor Cyan
Write-Host "检查网易云登录状态"
$authStatus = Invoke-RestMethod -Uri "$baseUrl/api/auth/status" -Method GET
Write-Host "已登录: $($authStatus.logged_in), 用户: $($authStatus.user.nickname)" -ForegroundColor Green

Write-Host "`n下载测试歌曲 (songID: 2751665002)"
$headers = @{Authorization="Bearer $token"}
$downloadBody = @{song_id=2751665002; quality="high"} | ConvertTo-Json
$downloadResult = Invoke-RestMethod -Uri "$baseUrl/api/download/song" -Method POST -Body $downloadBody -ContentType "application/json" -Headers $headers
Write-Host "任务已创建: task_id = $($downloadResult.task_id)" -ForegroundColor Green

Write-Host "`n等待 5 秒后查看任务状态"
Start-Sleep -Seconds 5
$tasks = Invoke-RestMethod -Uri "$baseUrl/api/tasks" -Method GET -Headers $headers
Write-Host "任务状态:" -ForegroundColor Yellow
$tasks | ConvertTo-Json -Depth 3

Write-Host "`n9. 查看下载历史" -ForegroundColor Cyan
$history = Invoke-RestMethod -Uri "$baseUrl/api/download/history" -Method GET -Headers $headers
Write-Host "下载记录数: $($history.Count)"
if ($history.Count -gt 0) {
    Write-Host "最新下载:" -ForegroundColor Yellow
    $history[0] | Select-Object song_name, artist, file_path, status | ConvertTo-Json
}

Write-Host "`n测试完成" -ForegroundColor Cyan
