#!/bin/bash

REMOTE_USER="root"
REMOTE_HOST="192.241.142.53"
REMOTE_DIR="/root/vedh/"

# rsync -avz --progress ./ "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"
git ls-files | rsync -av --progress --files-from=- ./ "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"

echo ""
echo "↘️↖️ Sync completed ✅"

# raw script:
# rsync -avz --exclude-from=<(git ls-files --others --exclude-standard --ignored -o --directory) ./ root@192.241.142.53:/root/vedh/