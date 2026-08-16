# 测试断点续传和假开关功能

$baseUrl = "http://192.168.1.70:33550"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "1. 登录系统" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$loginBody = @{
    username = "89913969@qq.com"
    password = "Luo122115"
} | ConvertTo-Json

try {
    $loginResult = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "登录成功" -ForegroundColor Green
    $loginResult | ConvertTo-Json
    $token = $loginResult.token
    Write-Host "Token: $($token.Substring(0, [Math]::Min(50, $token.Length)))..." -ForegroundColor Green
} catch {
    Write-Host "登录失败: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "2. 获取当前设置" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "当前设置:" -ForegroundColor Yellow
$settings.settings | Select-Object resume_downloads, delete_removed, sync_mode, quality, song_format | ConvertTo-Json

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "3. 测试 resume_downloads 开关" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n3.1 设置 resume_downloads = false" -ForegroundColor Yellow
$body = '{"resume_downloads":"false"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n3.2 验证设置" -ForegroundColor Yellow
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "resume_downloads = $($settings.settings.resume_downloads)" -ForegroundColor Green

Write-Host "`n3.3 设置 resume_downloads = true" -ForegroundColor Yellow
$body = '{"resume_downloads":"true"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n3.4 验证设置" -ForegroundColor Yellow
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "resume_downloads = $($settings.settings.resume_downloads)" -ForegroundColor Green

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "4. 测试 delete_removed 开关" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n4.1 设置 delete_removed = true" -ForegroundColor Yellow
$body = '{"delete_removed":"true"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n4.2 验证设置" -ForegroundColor Yellow
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "delete_removed = $($settings.settings.delete_removed)" -ForegroundColor Green

Write-Host "`n4.3 恢复 delete_removed = false" -ForegroundColor Yellow
$body = '{"delete_removed":"false"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "5. 测试 sync_mode 定时模式" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n5.1 设置 sync_mode = schedule" -ForegroundColor Yellow
$body = '{"sync_mode":"schedule"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n5.2 设置 sync_weekdays 和 sync_time" -ForegroundColor Yellow
$body = '{"sync_weekdays":"[1,3,5]","sync_time":"08:30"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n5.3 验证设置" -ForegroundColor Yellow
$settings = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET
Write-Host "sync_mode = $($settings.settings.sync_mode)" -ForegroundColor Green
Write-Host "sync_weekdays = $($settings.settings.sync_weekdays)" -ForegroundColor Green
Write-Host "sync_time = $($settings.settings.sync_time)" -ForegroundColor Green

Write-Host "`n5.4 恢复 sync_mode = interval" -ForegroundColor Yellow
$body = '{"sync_mode":"interval"}'
Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $body -ContentType "application/json" | ConvertTo-Json

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "6. 测试下载功能（断点续传）" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n6.1 检查网易云登录状态" -ForegroundColor Yellow
try {
    $authStatus = Invoke-RestMethod -Uri "$baseUrl/api/auth/status" -Method GET
    Write-Host "网易云登录状态:" -ForegroundColor Green
    $authStatus | ConvertTo-Json
} catch {
    Write-Host "获取登录状态失败: $_" -ForegroundColor Yellow
}

Write-Host "`n6.2 下载测试歌曲（songID: 2751665002）" -ForegroundColor Yellow
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
        Write-Host "下载任务已创建:" -ForegroundColor Green
        $downloadResult | ConvertTo-Json
    } catch {
        Write-Host "下载请求失败: $_" -ForegroundColor Red
    }
} else {
    Write-Host "跳过：未获取到 token" -ForegroundColor Yellow
}

Write-Host "`n6.3 等待 3 秒后查看任务状态" -ForegroundColor Yellow
Start-Sleep -Seconds 3

if ($token) {
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        $tasks = Invoke-RestMethod -Uri "$baseUrl/api/tasks" -Method GET -Headers $headers
        Write-Host "当前任务:" -ForegroundColor Green
        $tasks | ConvertTo-Json -Depth 5
    } catch {
        Write-Host "获取任务状态失败: $_" -ForegroundColor Red
    }
}

Write-Host "`n6.4 再等待 5 秒后查看进度" -ForegroundColor Yellow
Start-Sleep -Seconds 5

if ($token) {
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        $tasks = Invoke-RestMethod -Uri "$baseUrl/api/tasks" -Method GET -Headers $headers
        Write-Host "任务进度:" -ForegroundColor Green
        $tasks | ConvertTo-Json -Depth 5
    } catch {
        Write-Host "获取任务状态失败: $_" -ForegroundColor Red
    }
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "7. 验证下载路径隔离" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n7.1 查看下载历史" -ForegroundColor Yellow
if ($token) {
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        $history = Invoke-RestMethod -Uri "$baseUrl/api/download/history" -Method GET -Headers $headers
        Write-Host "下载历史:" -ForegroundColor Green
        $history | ConvertTo-Json -Depth 5
    } catch {
        Write-Host "获取下载历史失败: $_" -ForegroundColor Red
    }
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "测试完成" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
