#!/bin/sh
# 生成 bcrypt 密码哈希并更新数据库
# 密码: Luo122115

# 使用 htpasswd 生成 bcrypt 哈希（如果可用）或直接用 Python
if command -v htpasswd >/dev/null 2>&1; then
    HASH=$(htpasswd -nbBC 10 "" "Luo122115" | tr -d ':\n')
elif command -v python3 >/dev/null 2>&1; then
    HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw(b'Luo122115', bcrypt.gensalt()).decode())")
else
    echo "No bcrypt tool available"
    exit 1
fi

echo "Generated hash: $HASH"

# 更新数据库
sqlite3 /data/db/netmusic.db "UPDATE system_users SET password_hash='$HASH' WHERE username='yinluo'"
echo "Password updated for user: yinluo"

# 验证
sqlite3 /data/db/netmusic.db "SELECT username, substr(password_hash, 1, 10) as hash_prefix FROM system_users WHERE username='yinluo'"
