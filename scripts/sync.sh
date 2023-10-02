#!/bin/bash

# Variables
REMOTE_USER="root"
REMOTE_HOST="www.edhgo.com"
REMOTE_DIR="/root/edh-go/"

# Rsync command
rsync -avz --progress ./ "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"

echo "Sync completed."
