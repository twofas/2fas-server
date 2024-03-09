package config

import "time"

type PassConfig struct {
	Addr                                string        `envconfig:"PASS_ADDR" default:":8082"`
	KMSKeyID                            string        `envconfig:"KMS_KEY_ID" default:"alias/pass_service_signing_key"`
	AWSEndpoint                         string        `envconfig:"AWS_ENDPOINT" default:""`
	AWSRegion                           string        `envconfig:"AWS_REGION" default:"us-east-2"`
	FakeMobilePush                      bool          `envconfig:"FAKE_MOBILE_PUSH" default:"false"`
	PairingRequestTokenValidityDuration time.Duration `envconfig:"PAIRING_REQUEST_TOKEN_VALIDITY_DURATION" default:"8765h"` // 1 year
}
