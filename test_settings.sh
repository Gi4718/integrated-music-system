#!/bin/bash

# 测试设置保存功能
echo "=== 测试自动同步开关保存 ==="

# 1. 获取当前设置
echo "1. 当前设置:"
curl -s http://localhost:33550/api/settings | jq '.settings.auto_sync'

# 2. 尝试保存设置（使用正确的JSON格式）
echo -e "\n2. 尝试保存 auto_sync=true:"
curl -s -X POST http://localhost:33550/api/settings \
  -H 'Content-Type: application/json' \
  -d '{"auto_sync":"true"}'

# 3. 再次获取设置
echo -e "\n\n3. 保存后的设置:"
curl -s http://localhost:33550/api/settings | jq '.settings.auto_sync'

# 4. 尝试保存为false
echo -e "\n4. 尝试保存 auto_sync=false:"
curl -s -X POST http://localhost:33550/api/settings \
  -H 'Content-Type: application/json' \
  -d '{"auto_sync":"false"}'

# 5. 验证
echo -e "\n\n5. 最终设置:"
curl -s http://localhost:33550/api/settings | jq '.settings.auto_sync'

echo -e "\n=== 测试完成 ==="
