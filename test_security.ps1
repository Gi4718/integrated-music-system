# 安全漏洞修复测试脚本

$baseUrl = "http://192.168.1.70:33550"
$testUser = "yinluo"
$testPassword = "Luo122115"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "安全漏洞修复测试" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 1. 测试登录功能（bcrypt密码哈希）
Write-Host "`n[1] 测试登录功能（bcrypt密码哈希）" -ForegroundColor Yellow
$loginBody = @{
    username = $testUser
    password = $testPassword
} | ConvertTo-Json

try {
    $loginResult = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "✓ 登录成功" -ForegroundColor Green
    $token = $loginResult.token
    Write-Host "  Token: $($token.Substring(0, [Math]::Min(50, $token.Length)))..." -ForegroundColor Gray
} catch {
    Write-Host "✗ 登录失败: $_" -ForegroundColor Red
    exit 1
}

# 2. 测试SQL注入防护
Write-Host "`n[2] 测试SQL注入防护" -ForegroundColor Yellow
$sqlInjectionTests = @(
    @{ name = "单引号注入"; payload = "' OR '1'='1" },
    @{ name = "双引号注入"; payload = '" OR "1"="1' },
    @{ name = "注释注入"; payload = "admin' --" },
    @{ name = "UNION注入"; payload = "' UNION SELECT * FROM users --" }
)

foreach ($test in $sqlInjectionTests) {
    Write-Host "  测试: $($test.name)" -ForegroundColor Gray
    $injectBody = @{
        username = $test.payload
        password = $testPassword
    } | ConvertTo-Json
    
    try {
        $result = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $injectBody -ContentType "application/json"
        Write-Host "  ✗ 危险! SQL注入可能成功" -ForegroundColor Red
    } catch {
        if ($_.Exception.Response.StatusCode -eq 401) {
            Write-Host "  ✓ SQL注入被阻止" -ForegroundColor Green
        } else {
            Write-Host "  ? 异常响应: $_" -ForegroundColor Yellow
        }
    }
}

# 3. 测试JWT Token验证
Write-Host "`n[3] 测试JWT Token验证" -ForegroundColor Yellow

# 3.1 测试无效Token
Write-Host "  测试无效Token" -ForegroundColor Gray
try {
    $headers = @{ "Authorization" = "Bearer invalid_token_here" }
    $result = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET -Headers $headers
    Write-Host "  ✗ 危险! 无效Token被接受" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "  ✓ 无效Token被拒绝" -ForegroundColor Green
    } else {
        Write-Host "  ? 异常响应: $_" -ForegroundColor Yellow
    }
}

# 3.2 测试过期Token（需要手动创建）
Write-Host "  测试过期Token" -ForegroundColor Gray
# 这里跳过，因为需要特殊处理

# 3.3 测试有效Token
Write-Host "  测试有效Token" -ForegroundColor Gray
try {
    $headers = @{ "Authorization" = "Bearer $token" }
    $result = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET -Headers $headers
    Write-Host "  ✓ 有效Token被接受" -ForegroundColor Green
} catch {
    Write-Host "  ✗ 有效Token被拒绝: $_" -ForegroundColor Red
}

# 4. 测试权限控制（全局设置）
Write-Host "`n[4] 测试权限控制" -ForegroundColor Yellow

# 4.1 普通用户尝试修改全局设置
Write-Host "  测试普通用户修改全局设置" -ForegroundColor Gray
$globalSettingBody = @{
    "ssl_mode" = "none"
} | ConvertTo-Json

try {
    $headers = @{ "Authorization" = "Bearer $token" }
    $result = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method POST -Body $globalSettingBody -ContentType "application/json" -Headers $headers
    if ($result.error) {
        Write-Host "  ✓ 权限控制生效: $($result.error)" -ForegroundColor Green
    } else {
        Write-Host "  ? 响应: $($result | ConvertTo-Json)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  ? 异常: $_" -ForegroundColor Yellow
}

# 5. 测试路径遍历防护
Write-Host "`n[5] 测试路径遍历防护" -ForegroundColor Yellow

$pathTraversalTests = @(
    @{ name = "绝对路径"; path = "/etc/passwd" },
    @{ name = "相对路径"; path = "../../etc/passwd" },
    @{ name = "编码路径"; path = "%2e%2e%2fetc%2fpasswd" }
)

foreach ($test in $pathTraversalTests) {
    Write-Host "  测试: $($test.name)" -ForegroundColor Gray
    $validateBody = @{
        cert_path = $test.path
        key_path = $test.path
    } | ConvertTo-Json
    
    try {
        $headers = @{ "Authorization" = "Bearer $token" }
        $result = Invoke-RestMethod -Uri "$baseUrl/api/ssl/validate" -Method POST -Body $validateBody -ContentType "application/json" -Headers $headers
        if ($result.valid -eq $false -and $result.message -like "*必须在 /data/ssl/ 目录下*") {
            Write-Host "  ✓ 路径遍历被阻止" -ForegroundColor Green
        } else {
            Write-Host "  ? 响应: $($result | ConvertTo-Json)" -ForegroundColor Yellow
        }
    } catch {
        if ($_.Exception.Response.StatusCode -eq 400) {
            Write-Host "  ✓ 路径遍历被阻止" -ForegroundColor Green
        } else {
            Write-Host "  ? 异常: $_" -ForegroundColor Yellow
        }
    }
}

# 6. 测试登录错误信息泄露
Write-Host "`n[6] 测试登录错误信息泄露" -ForegroundColor Yellow

# 6.1 错误用户名
Write-Host "  测试错误用户名" -ForegroundColor Gray
$wrongUserBody = @{
    username = "wronguser"
    password = $testPassword
} | ConvertTo-Json

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $wrongUserBody -ContentType "application/json"
    Write-Host "  ✗ 危险! 错误用户名被接受" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        $errorDetail = $_.ErrorDetails.Message | ConvertFrom-Json
        if ($errorDetail.error -eq "用户名或密码错误") {
            Write-Host "  ✓ 错误信息已统一" -ForegroundColor Green
        } else {
            Write-Host "  ✗ 危险! 错误信息泄露: $($errorDetail.error)" -ForegroundColor Red
        }
    }
}

# 6.2 错误密码
Write-Host "  测试错误密码" -ForegroundColor Gray
$wrongPassBody = @{
    username = $testUser
    password = "wrongpassword"
} | ConvertTo-Json

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/api/system/login" -Method POST -Body $wrongPassBody -ContentType "application/json"
    Write-Host "  ✗ 危险! 错误密码被接受" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        $errorDetail = $_.ErrorDetails.Message | ConvertFrom-Json
        if ($errorDetail.error -eq "用户名或密码错误") {
            Write-Host "  ✓ 错误信息已统一" -ForegroundColor Green
        } else {
            Write-Host "  ✗ 危险! 错误信息泄露: $($errorDetail.error)" -ForegroundColor Red
        }
    }
}

# 7. 测试CORS配置
Write-Host "`n[7] 测试CORS配置" -ForegroundColor Yellow
try {
    $headers = @{
        "Origin" = "http://evil.com"
        "Authorization" = "Bearer $token"
    }
    $result = Invoke-RestMethod -Uri "$baseUrl/api/settings" -Method GET -Headers $headers
    Write-Host "  ? CORS配置已启用（需要检查响应头）" -ForegroundColor Yellow
} catch {
    Write-Host "  ? CORS测试异常: $_" -ForegroundColor Yellow
}

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "测试完成" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
