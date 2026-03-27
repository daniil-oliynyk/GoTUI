package config

type ChatClientConfig struct {
	APIKey string `env:"API_KEY"`
	Model  string
}
