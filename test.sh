#!/bin/bash

cd /home/renato/projetos/mygoprojs/src/github.com/renatocron/etamu-file2tar/
mkdir -p ./tmp/work-dir/
mkdir -p ./tmp/base-dir/

rm -rfv ./tmp/base-dir/*

. test-upload.sh &
. test-upload.sh

while true; do
  wait -n || {
    code="$?"
    ([[ $code = "127" ]] && exit 0 || exit "$code")
    break
  }
done;

curl 'localhost:1323/debug'
