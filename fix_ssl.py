import sqlite3

conn = sqlite3.connect('/tmp/netmusic.db')
conn.execute("UPDATE settings SET value='false' WHERE key='ssl_redirect'")
conn.execute("UPDATE settings SET value='none' WHERE key='ssl_mode'")
conn.commit()
conn.close()
print('done')
