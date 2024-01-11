#!/bin/sh

apt install --assume-yes jq

AWS_REGION=us-east-1
KEY_ALIAS=pass_service

response=$(awslocal kms create-key \
  --region $AWS_REGION \
  --key-usage SIGN_VERIFY \
  --customer-master-key-spec ECC_NIST_P256)

key_id=$(echo "${response}" | jq -r '.KeyMetadata.KeyId')

awslocal kms create-alias \
  --region $AWS_REGION \
  --alias-name "alias/$KEY_ALIAS" \
  --target-key-id "${key_id}"
