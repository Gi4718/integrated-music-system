#!/bin/sh
sqlite3 /data/db/netmusic.db "UPDATE system_users SET locked_until=NULL, failed_attempts=0 WHERE username='yinluo';"
sqlite3 /data/db/netmusic.db "SELECT * FROM system_users WHERE username='yinluo';"
