package config

type PassConfig struct {
	Addr        string `envconfig:"PASS_ADDR" default:":8082"`
	KMSKeyID    string `envconfig:"KMS_KEY_ID" default:"alias/pass_service_signing_key"`
	AWSEndpoint string `envconfig:"AWS_ENDPOINT" default:""`
}
