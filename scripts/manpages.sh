#!/bin/sh
# scripts/manpages.sh
set -e
rm -rf manpages
mkdir manpages
go run . man | gzip -c -9 > manpages/papercrypt.1.gz
