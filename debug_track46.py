import urllib.request
import json

print("=== 诊断第46首歌数据 ===")
print()

# 获取歌单详情
r = urllib.request.urlopen('http://127.0.0.1:3000/playlist/detail?id=3136952023')
d = json.loads(r.read())
p = d.get('playlist', {})
tracks = p.get('tracks', [])
tids = p.get('trackIds', [])

print(f'tracks数量: {len(tracks)}')
print(f'trackIds数量: {len(tids)}')
print()

# 检查第45-47首
for i in [44, 45, 46]:
    if i < len(tracks):
        t = tracks[i]
        print(f'=== tracks[{i}] (第{i+1}首) ===')
        print(f'id: {t.get("id")}')
        print(f'name: "{t.get("name")}" (type: {type(t.get("name")).__name__})')
        print(f'dt: {t.get("dt")}')
        print(f'tns: {t.get("tns")}')
        print(f'alia: {t.get("alia")}')
        print(f'ar: {t.get("ar")}')
        print(f'al: {t.get("al")}')
        print()

# 测试批量歌曲详情API
if len(tracks) > 45:
    id46 = tracks[45].get('id')
    print(f'=== 测试 song/detail?ids={id46} ===')
    r2 = urllib.request.urlopen(f'http://127.0.0.1:3000/song/detail?ids={id46}')
    d2 = json.loads(r2.read())
    songs = d2.get('songs', [])
    print(f'songs数量: {len(songs)}')
    if songs:
        s = songs[0]
        print(f'id: {s.get("id")}')
        print(f'name: "{s.get("name")}" (type: {type(s.get("name")).__name__})')
        print(f'ALL KEYS: {list(s.keys())}')
        print(f'FULL DATA: {json.dumps(s, ensure_ascii=False)[:500]}')
