#!/bin/bash

# 调试歌单详情 API

PLAYLIST_ID=${1:-2829816518}

echo "=== 测试歌单 ID: $PLAYLIST_ID ==="
echo ""

# 测试后端 API
echo "1. 测试后端 API (/api/playlist/detail):"
curl -s "http://192.168.1.70:33550/api/playlist/detail?id=$PLAYLIST_ID" | jq '{
  playlist_exists: (.playlist != null),
  tracks_count: (.tracks | length),
  first_track: .tracks[0],
  trackIds_count: (.playlist.trackIds | length),
  tracks_type: (.playlist.tracks | type),
  tracks_count_raw: (.playlist.tracks | length)
}'

echo ""
echo "2. 直接测试网易云 API (/playlist/detail):"
curl -s "http://127.0.0.1:3000/playlist/detail?id=$PLAYLIST_ID" | jq '{
  code: .code,
  playlist_exists: (.playlist != null),
  trackIds_count: (.playlist.trackIds | length),
  tracks_type: (.playlist.tracks | type),
  tracks_count: (.playlist.tracks | length),
  first_track: .playlist.tracks[0]
}'
