#!/bin/bash
# 注册系统用户并登录

BASE_URL="http://localhost:33550"

echo "1. 注册系统用户"
REGISTER_RESULT=$(curl -s -X POST "$BASE_URL/api/system/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"89913969@qq.com","password":"Luo122115"}')
echo "$REGISTER_RESULT"

echo ""
echo "2. 登录获取 token"
LOGIN_RESULT=$(curl -s -X POST "$BASE_URL/api/system/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"89913969@qq.com","password":"Luo122115"}')
echo "$LOGIN_RESULT"

TOKEN=$(echo "$LOGIN_RESULT" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "登录失败，无法获取 token"
  exit 1
fi

echo ""
echo "Token: ${TOKEN:0:50}..."

echo ""
echo "3. 测试下载功能（断点续传）"
echo "3.1 先设置 resume_downloads = true"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"resume_downloads":"true"}'
echo ""

echo "3.2 下载测试歌曲（songID: 2751665002）"
DOWNLOAD_RESULT=$(curl -s -X POST "$BASE_URL/api/download/song" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"song_id":2751665002,"quality":"high"}')
echo "$DOWNLOAD_RESULT"

echo ""
echo "4. 查看任务状态"
curl -s "$BASE_URL/api/tasks" -H "Authorization: Bearer $TOKEN" | head -c 500
echo ""

echo ""
echo "5. 等待 5 秒后再次查看进度"
sleep 5
curl -s "$BASE_URL/api/tasks" -H "Authorization: Bearer $TOKEN" | head -c 500
echo ""

echo ""
echo "6. 测试设置 resume_downloads = false"
curl -s -X POST "$BASE_URL/api/settings" \
  -H "Content-Type: application/json" \
  -d '{"resume_downloads":"false"}'
echo ""

echo ""
echo "7. 下载另一首歌曲测试（songID: 1341452231）"
DOWNLOAD_RESULT2=$(curl -s -X POST "$BASE_URL/api/download/song" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"song_id":1341452231,"quality":"high"}')
echo "$DOWNLOAD_RESULT2"

echo ""
echo "8. 查看任务状态"
curl -s "$BASE_URL/api/tasks" -H "Authorization: Bearer $TOKEN" | head -c 800
echo ""

echo ""
echo "9. 验证 delete_removed 设置"
curl -s "$BASE_URL/api/settings" | grep -o '"delete_removed":"[^"]*"'
echo ""

echo ""
echo "10. 验证 sync_mode 设置"
curl -s "$BASE_URL/api/settings" | grep -o '"sync_mode":"[^"]*"'
echo ""

echo ""
echo "测试完成！"
