#!/bin/bash

cd /home/renato/projetos/mygoprojs/src/github.com/renatocron/etamu-file2tar/

rm -rfv ./tmp/base-dir/*

echo -n "aaa" > ./tmp/base-dir/file-a
echo -n "b" > ./tmp/base-dir/file-b

export NOW=$(printf '%(%s)T\n' -1)
curl -X POST "localhost:1323/add?name=file-a&name=file-b&camera_id=123&timestamp=$NOW" | json_xs
