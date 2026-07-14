#!/bin/sh
echo "=== Playlist Detail ==="
wget -qO- 'http://127.0.0.1:3000/playlist/detail?id=3136952023' > /tmp/playlist.json
python3 -c '
import json
with open("/tmp/playlist.json") as f:
    d = json.load(f)
p = d.get("playlist", {})
tracks = p.get("tracks", [])
tids = p.get("trackIds", [])
print(f"tracks: {len(tracks)}, trackIds: {len(tids)}")
for i, t in enumerate(tracks):
    n = t.get("name")
    if n is None or n == "":
        print(f"  EMPTY track[{i}] id={t.get(chr(34)+\"id\"+chr(34))}")
' 2>&1 || echo "python parse failed"

echo ""
echo "=== Song Detail Batch ==="
wget -qO- 'http://127.0.0.1:3000/song/detail?ids=1959845600' > /tmp/song.json
python3 -c '
import json
with open("/tmp/song.json") as f:
    d = json.load(f)
for s in d.get("songs", []):
    print(f"id={s.get(\"id\")} name={repr(s.get(\"name\"))} ar={s.get(\"ar\")} al={s.get(\"al\")} dt={s.get(\"dt\")}")
' 2>&1 || echo "python parse failed"

echo ""
echo "=== Song URL ==="
wget -qO- 'http://127.0.0.1:3000/song/url/v1?id=1959845600&br=320000' > /tmp/url.json
python3 -c '
import json
with open("/tmp/url.json") as f:
    d = json.load(f)
for s in d.get("data", []):
    print(f"id={s.get(\"id\")} url={s.get(\"url\")} code={s.get(\"code\")} size={s.get(\"size\")} br={s.get(\"br\")}")
' 2>&1 || echo "python parse failed"
