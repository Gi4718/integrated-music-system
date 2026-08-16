#!/bin/bash

# 安全漏洞修复测试脚本

BASE_URL="http://localhost:33550"
TEST_USER="yinluo"
TEST_PASSWORD="Luo122115"

echo "=========================================="
echo "安全漏洞修复测试"
echo "=========================================="

# 1. 测试登录功能（bcrypt密码哈希）
echo ""
echo "[1] 测试登录功能（bcrypt密码哈希）"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/system/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASSWORD\"}")

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo "✓ 登录成功"
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "  Token: ${TOKEN:0:50}..."
else
    echo "✗ 登录失败: $LOGIN_RESPONSE"
    exit 1
fi

# 2. 测试SQL注入防护
echo ""
echo "[2] 测试SQL注入防护"

SQL_TESTS=(
    "' OR '1'='1"
    '" OR "1"="1'
    "admin' --"
    "' UNION SELECT * FROM users --"
)

for payload in "${SQL_TESTS[@]}"; do
    echo "  测试: $payload"
    RESPONSE=$(curl -s -X POST "$BASE_URL/api/system/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\":\"$payload\",\"password\":\"$TEST_PASSWORD\"}")
    
    if echo "$RESPONSE" | grep -q "401"; then
        echo "  ✓ SQL注入被阻止"
    else
        echo "  ✗ 危险! SQL注入可能成功: $RESPONSE"
    fi
done

# 3. 测试JWT Token验证
echo ""
echo "[3] 测试JWT Token验证"

echo "  测试无效Token"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/settings" \
  -H "Authorization: Bearer invalid_token_here")

if echo "$RESPONSE" | grep -q "401"; then
    echo "  ✓ 无效Token被拒绝"
else
    echo "  ✗ 危险! 无效Token被接受: $RESPONSE"
fi

echo "  测试有效Token"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/settings" \
  -H "Authorization: Bearer $TOKEN")

if echo "$RESPONSE" | grep -q "settings"; then
    echo "  ✓ 有效Token被接受"
else
    echo "  ✗ 有效Token被拒绝: $RESPONSE"
fi

# 4. 测试权限控制（全局设置）
echo ""
echo "[4] 测试权限控制"

echo "  测试普通用户修改全局设置"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ssl_mode":"none"}')

if echo "$RESPONSE" | grep -q "只有管理员"; then
    echo "  ✓ 权限控制生效"
else
    echo "  ? 响应: $RESPONSE"
fi

# 5. 测试路径遍历防护
echo ""
echo "[5] 测试路径遍历防护"

PATH_TESTS=(
    "/etc/passwd"
    "../../etc/passwd"
    "%2e%2e%2fetc%2fpasswd"
)

for path in "${PATH_TESTS[@]}"; do
    echo "  测试: $path"
    RESPONSE=$(curl -s -X POST "$BASE_URL/api/ssl/validate" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{\"cert_path\":\"$path\",\"key_path\":\"$path\"}")
    
    if echo "$RESPONSE" | grep -q "必须在 /data/ssl/ 目录下"; then
        echo "  ✓ 路径遍历被阻止"
    else
        echo "  ? 响应: $RESPONSE"
    fi
done

# 6. 测试登录错误信息泄露
echo ""
echo "[6] 测试登录错误信息泄露"

echo "  测试错误用户名"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/system/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"wronguser","password":"'$TEST_PASSWORD'"}')

if echo "$RESPONSE" | grep -q "用户名或密码错误"; then
    echo "  ✓ 错误信息已统一"
else
    echo "  ✗ 危险! 错误信息泄露: $RESPONSE"
fi

echo "  测试错误密码"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/system/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"'$TEST_USER'","password":"wrongpassword"}')

if echo "$RESPONSE" | grep -q "用户名或密码错误"; then
    echo "  ✓ 错误信息已统一"
else
    echo "  ✗ 危险! 错误信息泄露: $RESPONSE"
fi

# 7. 测试CORS配置
echo ""
echo "[7] 测试CORS配置"
RESPONSE=$(curl -s -I -X OPTIONS "$BASE_URL/api/settings" \
  -H "Origin: http://evil.com" \
  -H "Access-Control-Request-Method: GET")

if echo "$RESPONSE" | grep -q "Access-Control-Allow-Origin"; then
    echo "  ? CORS配置已启用"
else
    echo "  ? CORS测试无响应"
fi

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
