#!/usr/bin/env python3
"""Fix duplicate saveSyncSettings in Settings.vue"""
import sys

filepath = sys.argv[1]
with open(filepath, 'r') as f:
    content = f.read()

# Find the two saveSyncSettings definitions
# First one has updateSyncPlaylists call, second doesn't
# We need to keep the first one and remove the second

marker1 = "// 保存同步歌单配置\nconst saveSyncSettings = async () => {"
marker2 = "// 保存歌单同步配置\nconst saveSyncSettings = async () => {"

if marker1 in content and marker2 in content:
    # Find the second occurrence and remove it
    # The second one starts with "// 保存歌单同步配置"
    idx2 = content.index(marker2)
    # Find the end of the second function - look for the next "// 保存" or "const save"
    rest = content[idx2:]
    # Find the closing of the function - look for next function definition
    lines = rest.split('\n')
    removed_lines = 0
    for i, line in enumerate(lines):
        if i > 0 and (line.startswith('// 保存') or line.startswith('const save')):
            # Found next function, remove lines from idx2 to idx2+i
            content = content[:idx2] + content[idx2 + i * len('\n') + len('\n'):]
            removed_lines = i
            break

    with open(filepath, 'w') as f:
        f.write(content)
    print(f"OK: Removed duplicate saveSyncSettings ({removed_lines} lines)")
else:
    print(f"Markers found: marker1={marker1 in content}, marker2={marker2 in content}")
    print("Checking for other patterns...")
    count = content.count('const saveSyncSettings')
    print(f"Found {count} saveSyncSettings definitions")
