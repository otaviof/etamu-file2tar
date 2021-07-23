#!/bin/bash

cd /home/renato/projetos/mygoprojs/src/github.com/renatocron/etamu-file2tar/

BASE_DIR=./tmp/base-dir WORK_DIR=./tmp/work-dir go run -race *.go
