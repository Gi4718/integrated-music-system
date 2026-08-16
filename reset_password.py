#!/usr/bin/env python3
import sqlite3
import bcrypt

# 生成 bcrypt 密码哈希
password = "Luo122115"
hashed = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt()).decode('utf-8')

# 更新数据库
conn = sqlite3.connect('/data/db/netmusic.db')
cursor = conn.cursor()
cursor.execute("UPDATE system_users SET password_hash=?, failed_attempts=0, locked_until=NULL WHERE username='yinluo'", (hashed,))
conn.commit()

# 验证
cursor.execute("SELECT username, substr(password_hash,1,10) FROM system_users WHERE username='yinluo'")
result = cursor.fetchone()
print(f"Updated: {result}")
conn.close()
