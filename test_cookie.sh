#!/bin/sh
COOKIE=$(sqlite3 /data/db/netmusic.db 'SELECT cookie FROM users LIMIT 1;')
echo "Cookie length: $(echo -n "$COOKIE" | wc -c)"
echo ""
echo "=== user/account ==="
wget -qO- --header="Cookie: $COOKIE" "http://127.0.0.1:3000/user/account" 2>/dev/null | head -c 2000
echo ""
echo ""
echo "=== user/subcount ==="
wget -qO- --header="Cookie: $COOKIE" "http://127.0.0.1:3000/user/subcount" 2>/dev/null | head -c 1000
echo ""
echo ""
echo "=== login/status ==="
wget -qO- --header="Cookie: $COOKIE" "http://127.0.0.1:3000/login/status" 2>/dev/null | head -c 2000
echo ""
