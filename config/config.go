package main

import (
	"github.com/caarlos0/env"
)

type Conf struct {
	DiscordToken         string `env:"DISCORD_TOKEN"`
	DiscordChannel       string `env:"DISCORD_CHANNEL"`
	NightbotClientID     string `env:"NIGHTBOT_CLIENT_ID"`
	NightbotRedirectURI  string `env:"NIGHTBOT_REDIRECT_URI"`
	NightbitAuthURL      string `env:"NIGHTBOT_AUTH_URL"`
	NightbotClientSecret string `env:"NIGHTBOT_CLIENT_SECRET"`
}

// Config is a global configuration that is used within qoqbot
var Config Conf

func init() {
	FromEnv()
}

// FromEnv returns the configurations found in the environment variables within docker-compose
func FromEnv() Conf {
	if err := env.Parse(&Config); err != nil {
		panic(err)
	}
	return Config
}
