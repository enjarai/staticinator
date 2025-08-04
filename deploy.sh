set -e

FOLDER=$INPUT_FOLDER
HOST=$INPUT_HOST
TOKEN=$INPUT_TOKEN
DEPLOY_ENDPOINT=$INPUT_DEPLOY_ENDPOINT

echo "Compressing input"

TMP_DIR=$(mktemp -d)
tar -czvf "$TMP_DIR/archive.tar.gz" -C "$FOLDER" .

echo "Sending to deploy endpoint"

curl -F "file=@$TMP_DIR/archive.tar.gz" -H "Token: $TOKEN" -H "Target-Host: $HOST" "$DEPLOY_ENDPOINT" -v

echo "Deployment complete"
