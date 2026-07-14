#!/bin/bash

# 调试歌单API返回数据

echo "=== 测试网易云API歌单详情 ==="
curl -s "http://127.0.0.1:3000/playlist/detail?id=2829816518" | jq '{
  code: .code,
  playlist_exists: (.playlist != null),
  playlist_name: .playlist.name,
  trackCount: .playlist.trackCount,
  tracks_type: (.playlist.tracks | type),
  tracks_length: (.playlist.tracks | length),
  trackIds_type: (.playlist.trackIds | type),
  trackIds_length: (.playlist.trackIds | length),
  first_track: .playlist.tracks[0],
  first_trackId: .playlist.trackIds[0]
}'

echo ""
echo "=== 测试后端API ==="
curl -s "http://127.0.0.1:33550/api/playlist/detail?id=2829816518" | jq '{
  tracks_count: (.tracks | length),
  playlist_name: .playlist.name,
  first_track: .tracks[0]
}'
