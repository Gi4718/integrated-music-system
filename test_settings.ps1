# 测试断点续传和假开关功能

$baseUrl = "http://192.168.1.70:33550"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "1. 检查系统用户状态" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
$checkResult = Invoke-RestMethod -Uri "$baseUrl/api/system/check" -Method GET
$checkResult | ConvertTo-Json

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "2. 登录获取 token" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 尝试登录
try {
    $loginBody = @{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json
    
    $loginResult = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $loginBody -ContentType "application/json"
    $loginResult | ConvertTo-Json
    $token = $loginResult.token
} catch {
    Write-Host "登录失败，尝试注册新用户..." -ForegroundColor Yellow
    
    # 注册新用户
    $registerBody = @{
        username = "testuser"
        password = "test123456"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/system/register" -Method POST -Body $registerBody -ContentType "application/json"
        
        # 重新登录
        $loginResult = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $loginBody -ContentType "application/json"
        $loginResult | ConvertTo-Json
        $token = $loginResult.token
    } catch {
        Write-Host "注册或登录失败: $_" -ForegroundColor Red
    }
}

if ($token) {
    Write-Host "`nToken: $($token.Substring(0, [Math]::Min(50, $token.Length)))..." -ForegroundColor Green
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "3. 获取当前设置" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
$settings.settings | Select-Object resume_downloads, delete_removed, sync_mode, sync_weekdays, sync_time, quality, song_format | ConvertTo-Json

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "4. 测试 resume_downloads 开关" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n4.1 设置 resume_downloads = false"
$body = '{"resume_downloads":"false"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n4.2 验证设置已保存"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "resume_downloads = $($settings.settings.resume_downloads)"

Write-Host "`n4.3 设置 resume_downloads = true"
$body = '{"resume_downloads":"true"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n4.4 验证设置已保存"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "resume_downloads = $($settings.settings.resume_downloads)"

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "5. 测试 delete_removed 开关" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n5.1 设置 delete_removed = true"
$body = '{"delete_removed":"true"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n5.2 验证设置已保存"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "delete_removed = $($settings.settings.delete_removed)"

Write-Host "`n5.3 设置 delete_removed = false"
$body = '{"delete_removed":"false"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n5.4 验证设置已保存"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "delete_removed = $($settings.settings.delete_removed)"

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "6. 测试 sync_mode 定时模式" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n6.1 设置 sync_mode = schedule"
$body = '{"sync_mode":"schedule"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n6.2 设置 sync_weekdays 和 sync_time"
$body = '{"sync_weekdays":"[1,3,5]","sync_time":"08:30"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n6.3 验证设置已保存"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
$settings.settings | Select-Object sync_mode, sync_weekdays, sync_time | ConvertTo-Json

Write-Host "`n6.4 恢复 sync_mode = interval"
$body = '{"sync_mode":"interval"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "7. 测试用户隔离设置" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n7.1 设置 quality = lossless"
$body = '{"quality":"lossless"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n7.2 验证 quality 设置"
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "quality = $($settings.settings.quality)"

Write-Host "`n7.3 恢复 quality = high"
$body = '{"quality":"high"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "8. 测试下载功能（需要网易云 cookie）" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n8.1 检查网易云登录状态"
try {
    $authStatus = Invoke-RestMethod -Uri "$baseUrl/api/auth/status" -Method GET
    $authStatus | ConvertTo-Json
} catch {
    Write-Host "获取登录状态失败: $_" -ForegroundColor Yellow
}

Write-Host "`n8.2 尝试下载一首测试歌曲（songID: 2751665002）"
if ($token) {
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        $downloadBody = @{
            song_id = 2751665002
            quality = "high"
        } | ConvertTo-Json
        
        $downloadResult = Invoke-RestMethod -Uri "$baseUrl/api/download/song" -Method POST -Body $downloadBody -ContentType "application/json" -Headers $headers
        $downloadResult | ConvertTo-Json
    } catch {
        Write-Host "下载请求失败: $_" -ForegroundColor Red
    }
} else {
    Write-Host "跳过：未获取到 token" -ForegroundColor Yellow
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "9. 查看任务状态" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

if ($token) {
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        $tasks = Invoke-RestMethod -Uri "$baseUrl/api/tasks" -Method GET -Headers $headers
        $tasks | ConvertTo-Json -Depth 5
    } catch {
        Write-Host "获取任务状态失败: $_" -ForegroundColor Red
    }
} else {
    Write-Host "跳过：未获取到 token" -ForegroundColor Yellow
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "测试完成" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
