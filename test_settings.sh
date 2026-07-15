#!/bin/bash
echo '{"auto_sync":"false"}' > /tmp/test.json
echo "Request body:"
cat /tmp/test.json
echo ""
echo "Response:"
curl -s -X POST http://localhost:33550/api/settings -H 'Content-Type: application/json' -d @/tmp/test.json
