#!/bin/sh
# Check playlist detail API response
echo "=== Playlist Detail ==="
wget -qO- 'http://127.0.0.1:3000/playlist/detail?id=3136952023' | python3 -c '
import sys, json
d = json.load(sys.stdin)
p = d.get("playlist", {})
tracks = p.get("tracks", [])
tids = p.get("trackIds", [])
print(f"tracks count: {len(tracks)}")
print(f"trackIds count: {len(tids)}")
for i, t in enumerate(tracks):
    n = t.get("name")
    if n is None or n == "":
        print(f"  track[{i}] id={t.get(\"id\")} name={n}")
'

echo ""
echo "=== Song Detail for null-name song ==="
wget -qO- 'http://127.0.0.1:3000/song/detail?ids=1959845600' | python3 -c '
import sys, json
d = json.load(sys.stdin)
songs = d.get("songs", [])
for s in songs:
    print(f"id={s.get(\"id\")} name={s.get(\"name\")} ar={s.get(\"ar\")} al={s.get(\"al\")}")
'

echo ""
echo "=== Song URL for playback test ==="
wget -qO- 'http://127.0.0.1:3000/song/url/v1?id=1959845600&br=320000' | python3 -c '
import sys, json
d = json.load(sys.stdin)
data = d.get("data", [])
for s in data:
    print(f"id={s.get(\"id\")} url={s.get(\"url\")} code={s.get(\"code\")} size={s.get(\"size\")}")
'
