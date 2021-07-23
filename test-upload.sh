#!/bin/bash

NEW_UUID=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

echo -n "aaa" > ./tmp/base-dir/file-a$NEW_UUID
echo -n "b" > ./tmp/base-dir/file-b$NEW_UUID

export NOW=$(printf '%(%s)T\n' -1)
curl -X POST "localhost:1323/add?name=file-a$NEW_UUID&name=file-b$NEW_UUID&camera_id=123&timestamp=$NOW" | json_xs
