#!/bin/bash

REMOTE_USER="root"
REMOTE_HOST="www.edhgo.com"
REMOTE_DIR="/home/shakezula/edh-go"

# rsync -avz --progress ./ "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"
git ls-files | rsync -av --progress --files-from=- ./ "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"

echo ""
echo "↘️↖️ Sync completed ✅"
