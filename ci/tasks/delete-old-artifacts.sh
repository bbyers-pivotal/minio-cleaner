#!/usr/bin/env sh
set -e

#strip http/https
host="$(echo $host | sed 's~^http[s]*://~~')"

minio-cleaner -host=${host} -access-key=${access_key} -secret-key=${secret_key} -prefix=${prefix} -bucket=${bucket} -backups-to-keep=${backups_to_keep} -use-ssl=${use_ssl} -dry-run=${dry_run}