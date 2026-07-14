#!/bin/sh
# 从数据库获取 cookie
COOKIE=$(docker exec endfield-music sqlite3 /data/db/netmusic.db "SELECT cookie FROM users LIMIT 1;")

# 测试带 cookie 的 API 请求
docker exec endfield-music wget -qO- --header="Cookie: $COOKIE" 'http://127.0.0.1:3000/playlist/detail?id=2829816518' | grep -o '"tracks":\[' | head -c 100

echo ""
echo "Cookie length: $(echo -n "$COOKIE" | wc -c)"
