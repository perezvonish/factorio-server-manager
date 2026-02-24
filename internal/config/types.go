package config

type Config struct {
	Telegram       TelegramConfig
	FactorioServer FactorioServerConfig
	Docker         DockerConfig
}

type TelegramConfig struct {
	BotToken     string `env:"TELEGRAM_BOT_TOKEN" required:"true"`
	AllowedUsers string `env:"TELEGRAM_ALLOWED_USERS" envDefault:""`
}

type FactorioServerConfig struct {
	RconHost           string `env:"RCON_HOST" envDefault:"factorio"`
	RconPort           string `env:"RCON_PORT" envDefault:"27015"`
	GameHost           string `env:"FACTORIO_GAME_HOST" envDefault:"factorio"`
	GamePort           string `env:"FACTORIO_GAME_PORT" envDefault:"34197"`
	SavesDir           string `env:"FACTORIO_SAVES_DIR" envDefault:"/factorio/saves"`
	RconPwFile         string `env:"FACTORIO_RCON_PW_FILE" envDefault:"/factorio/config/rconpw"`
	ServerSettingsFile string `env:"FACTORIO_SERVER_SETTINGS_FILE" envDefault:"/factorio/config/server-settings.json"`
}

type DockerConfig struct {
	ContainerName string `env:"DOCKER_CONTAINER_NAME" envDefault:"factorio"`
}
