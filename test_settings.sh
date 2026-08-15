#!/bin/bash
# 测试断点续传和假开关功能

BASE_URL="http://192.168.1.70:33550"

echo "=========================================="
echo "1. 检查系统用户状态"
echo "=========================================="
curl -s "$BASE_URL/api/system/check" | jq .

echo ""
echo "=========================================="
echo "2. 登录获取 token"
echo "=========================================="
# 使用默认账号 admin/admin123 登录（如果存在）
LOGIN_RESULT=$(curl -s -X POST "$BASE_URL/api/system/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
echo "$LOGIN_RESULT" | jq .

TOKEN=$(echo "$LOGIN_RESULT" | jq -r '.token // empty')

if [ -z "$TOKEN" ]; then
  echo "登录失败，尝试注册新用户..."
  curl -s -X POST "$BASE_URL/api/system/register" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123456"}' | jq .
  
  LOGIN_RESULT=$(curl -s -X POST "$BASE_URL/api/system/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123456"}')
  echo "$LOGIN_RESULT" | jq .
  TOKEN=$(echo "$LOGIN_RESULT" | jq -r '.token')
fi

echo ""
echo "Token: ${TOKEN:0:50}..."

echo ""
echo "=========================================="
echo "3. 获取当前设置"
echo "=========================================="
curl -s "$BASE_URL/api/settings" | jq '.settings | {
  resume_downloads,
  delete_removed,
  sync_mode,
  sync_weekdays,
  sync_time,
  quality,
  song_format
}'

echo ""
echo "=========================================="
echo "4. 测试 resume_downloads 开关"
echo "=========================================="
echo "4.1 设置 resume_downloads = false"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"resume_downloads":"false"}' | jq .

echo ""
echo "4.2 验证设置已保存"
curl -s "$BASE_URL/api/settings" | jq '.settings.resume_downloads'

echo ""
echo "4.3 设置 resume_downloads = true"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"resume_downloads":"true"}' | jq .

echo ""
echo "4.4 验证设置已保存"
curl -s "$BASE_URL/api/settings" | jq '.settings.resume_downloads'

echo ""
echo "=========================================="
echo "5. 测试 delete_removed 开关"
echo "=========================================="
echo "5.1 设置 delete_removed = true"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"delete_removed":"true"}' | jq .

echo ""
echo "5.2 验证设置已保存"
curl -s "$BASE_URL/api/settings" | jq '.settings.delete_removed'

echo ""
echo "5.3 设置 delete_removed = false"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"delete_removed":"false"}' | jq .

echo ""
echo "5.4 验证设置已保存"
curl -s "$BASE_URL/api/settings" | jq '.settings.delete_removed'

echo ""
echo "=========================================="
echo "6. 测试 sync_mode 定时模式"
echo "=========================================="
echo "6.1 设置 sync_mode = schedule"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"sync_mode":"schedule"}' | jq .

echo ""
echo "6.2 设置 sync_weekdays 和 sync_time"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"sync_weekdays":"[1,3,5]","sync_time":"08:30"}' | jq .

echo ""
echo "6.3 验证设置已保存"
curl -s "$BASE_URL/api/settings" | jq '.settings | {sync_mode, sync_weekdays, sync_time}'

echo ""
echo "6.4 恢复 sync_mode = interval"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"sync_mode":"interval"}' | jq .

echo ""
echo "=========================================="
echo "7. 测试用户隔离设置"
echo "=========================================="
echo "7.1 设置 quality = lossless"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"quality":"lossless"}' | jq .

echo ""
echo "7.2 验证 quality 设置"
curl -s "$BASE_URL/api/settings" | jq '.settings.quality'

echo ""
echo "7.3 恢复 quality = high"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"quality":"high"}' | jq .

echo ""
echo "=========================================="
echo "8. 测试下载功能（需要网易云 cookie）"
echo "=========================================="
echo "8.1 检查网易云登录状态"
curl -s "$BASE_URL/api/auth/status" | jq .

echo ""
echo "8.2 尝试下载一首测试歌曲（songID: 2751665002）"
if [ -n "$TOKEN" ]; then
  curl -s -X POST "$BASE_URL/api/download/song" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"song_id":2751665002,"quality":"high"}' | jq .
else
  echo "跳过：未获取到 token"
fi

echo ""
echo "=========================================="
echo "9. 查看任务状态"
echo "=========================================="
if [ -n "$TOKEN" ]; then
  curl -s "$BASE_URL/api/tasks" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo "跳过：未获取到 token"
fi

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
