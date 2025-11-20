#!/usr/bin/env bash

URL="${1:-http://31.130.149.163:8080/team/add}"
TEAM_NAME="bench-$(date +%s)-$RANDOM"
USER1="${TEAM_NAME}-1"
USER2="${TEAM_NAME}-2"


PAYLOAD_FILE=$(mktemp)
cat <<JSON > "$PAYLOAD_FILE"
{
  "team_name": "${TEAM_NAME}",
  "members": [
    {"user_id": "${USER1}", "username": "${USER1}", "is_active": true},
    {"user_id": "${USER2}", "username": "${USER2}", "is_active": true}
  ]
}
JSON

echo "Running load test against ${URL} ..."
ab -n 8000 -c 10 -p "$PAYLOAD_FILE" -T application/json "$URL"
rm "$PAYLOAD_FILE"
