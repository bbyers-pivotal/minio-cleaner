#!/usr/bin/env bash
set -e

minio-cleaner -host=${host} -access-key=${access-key} -secret-key=${secret-key} -prefix=${prefix} -bucket=${bucket} -backups-to-keep=${backups-to-keep} -use-ssl=${use-ssl} -dry-run=${dry-run}