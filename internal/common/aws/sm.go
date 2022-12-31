package aws

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func getSecret(secretName string) (string, error) {
	sess, err := session.NewSession()

	if err != nil {
		return "", err
	}

	svc := secretsmanager.New(sess)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(input)

	if err != nil {
		return "", err
	}

	var secretString, decodedBinarySecret string

	if result.SecretString != nil {
		secretString = *result.SecretString

		return secretString, nil
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		length, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)

		if err != nil {
			return "", err
		}

		decodedBinarySecret = string(decodedBinarySecretBytes[:length])

		return decodedBinarySecret, nil
	}
}
