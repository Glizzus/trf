#!/bin/bash
set -e

function check_env_vars() {
    local vars=(
        "MINIO_URL"
        "MINIO_ROOT_USER"
        "MINIO_ROOT_PASSWORD"
        "MINISTRY_PASSWORD"
    )

    for var in "${vars[@]}"; do
        if [ -z "${!var}" ]; then
            echo "Environment variable $var is required"
            exit 1
        fi
    done
}

function set_minio_alias() {
    local retries=5
    until mc alias set minio "$MINIO_URL" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"; do
        retries=$((retries - 1))
        if [ $retries -eq 0 ]; then
            echo "Failed to set minio alias after 5 seconds"
            exit 1
        fi
        sleep 1
    done
}

check_env_vars
set_minio_alias

mc mb --ignore-existing minio/trf
# The bucket will be publicly readonly
mc anonymous set download minio/trf

cat <<EOF > /tmp/ministry-policy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetBucketLocation"
            ],
            "Resource": [
                "arn:aws:s3:::trf"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::trf/*"
            ]
        }
    ]
}
EOF

mc admin user add minio ministry "$MINISTRY_PASSWORD"
mc admin policy create minio ministry /tmp/ministry-policy.json

mc admin policy attach minio ministry --user ministry || true

exec "$@"
