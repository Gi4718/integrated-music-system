#!/bin/sh
curl -s -X POST http://localhost:33550/api/system/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"yinluo","password":"Luo122115"}'
echo ""
curl -s -X POST http://localhost:33550/api/system/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"yinluo","password":"wrongpass"}'
echo ""
