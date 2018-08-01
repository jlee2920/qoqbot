package config

import (
	"github.com/caarlos0/env"
)

// Conf holds the environment configurations taken from docker-compose.yml
type Conf struct {
	// Discord Information
	DiscordToken     string `env:"DISCORD_TOKEN"`
	DiscordURL       string `env:"DISCORD_URL"`
	DiscordBotID     string `env:"DISCORD_BOT_ID"`
	DiscordChannelID string `env:"DISCORD_CHANNEL_ID"`

	// Twitch Bot Information
	BotName     string `env:"BOT_NAME"`
	BotOAuth    string `env:"BOT_OAUTH"`
	ChannelName string `env:"CHANNEL_NAME"`

	// PostgreSQL Credentials
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBHost     string `env:"DB_HOST"`
	DBPort     int    `env:"DB_PORT"`
	DBName     string `env:"DB_NAME"`
}

// Config is a global configuration that is used within qoqbot
var Config Conf

// InitEnv grabs the environment variables found within the docker-compose.yml file
func InitEnv() Conf {
	// Config is a global configuration that is used within qoqbot
	if err := env.Parse(&Config); err != nil {
		panic(err)
	}
	return Config
}
