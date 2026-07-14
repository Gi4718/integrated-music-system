#!/bin/sh
echo "=== 诊断第46首歌数据 ==="
echo ""
echo "--- 1. 歌单详情API ---"
wget -qO- 'http://127.0.0.1:3000/playlist/detail?id=3136952023' > /tmp/playlist.json
python3 -c '
import json
with open("/tmp/playlist.json") as f:
    d = json.load(f)
p = d.get("playlist", {})
tracks = p.get("tracks", [])
tids = p.get("trackIds", [])
print(f"tracks数量: {len(tracks)}")
print(f"trackIds数量: {len(tids)}")
print()

# 检查前几首和45-47首
for i in [0, 1, 44, 45, 46]:
    if i < len(tracks):
        t = tracks[i]
        print(f"tracks[{i}]: id={t.get(\"id\")} name=\"{t.get(\"name\")}\" name_type={type(t.get(\"name\")).__name__} dt={t.get(\"dt\")}")
        if t.get("name") is None:
            print(f"  ALL KEYS: {list(t.keys())}")
            print(f"  FULL: {json.dumps(t, ensure_ascii=False)[:300]}")
'

echo ""
echo "--- 2. 批量歌曲详情API ---"
# 获取第46首的ID
ID46=$(python3 -c '
import json
with open("/tmp/playlist.json") as f:
    d = json.load(f)
p = d.get("playlist", {})
tracks = p.get("tracks", [])
tids = p.get("trackIds", [])
if len(tracks) > 45:
    print(tracks[45].get("id", ""))
elif len(tids) > 45:
    t = tids[45]
    if isinstance(t, dict):
        print(t.get("id", ""))
    else:
        print(t)
')
echo "第46首歌ID: $ID46"

if [ -n "$ID46" ]; then
    echo "查询 song/detail?ids=$ID46"
    wget -qO- "http://127.0.0.1:3000/song/detail?ids=$ID46" | python3 -c '
import sys, json
d = json.load(sys.stdin)
print(f"返回keys: {list(d.keys())}")
songs = d.get("songs", [])
print(f"songs数量: {len(songs)}")
for s in songs:
    print(f"  id={s.get(\"id\")} name=\"{s.get(\"name\")}\" name_type={type(s.get(\"name\")).__name__}")
    if s.get("name") is None:
        print(f"  ALL KEYS: {list(s.keys())}")
        print(f"  FULL: {json.dumps(s, ensure_ascii=False)[:500]}")
'
fi
