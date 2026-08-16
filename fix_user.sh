#!/bin/sh
# 删除 admin 用户，将 yinluo 改为 id=1
sqlite3 /data/db/netmusic.db <<EOF
DELETE FROM system_users WHERE username='admin';
UPDATE system_users SET id=1 WHERE username='yinluo';
SELECT id, username, substr(password_hash,1,10) FROM system_users;
EOF
