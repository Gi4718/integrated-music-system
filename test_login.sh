#!/bin/sh
BODY='{"username":"yinluo","password":"Luo122115"}'
RESULT=$(wget -qO- --post-data="$BODY" --header='Content-Type: application/json' http://localhost:33550/api/system/login 2>&1)
echo "Login result: $RESULT"
