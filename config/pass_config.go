package config

type PassConfig struct {
	Addr string `envconfig:"PASS_ADDR" default:":8084"`
}
