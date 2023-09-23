#!/bin/sh
# scripts/completions.sh
set -e
rm -rf completions
mkdir completions
for shell in bash zsh fish; do
  echo "Generating completions for $shell"
  go run . completion "$shell" > "completions/papercrypt.$shell"
done
