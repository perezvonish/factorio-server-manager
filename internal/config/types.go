package config

type Config struct {
	Telegram       TelegramConfig
	FactorioServer FactorioServerConfig
	Docker         DockerConfig
	WebApp         WebAppConfig
	ModPortal      ModPortalConfig
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

// WebAppConfig configures the built-in HTTP server for the save-upload Telegram WebApp.
type WebAppConfig struct {
	// Port — внутренний порт HTTP-сервера (по умолчанию 8080).
	Port string `env:"WEBAPP_PORT" envDefault:"8080"`
	// URL — публичный HTTPS-адрес, который открывается в Telegram WebApp.
	// Должен быть HTTPS, например: https://example.com:8080
	// Если не задан, /uploadSave шлёт текстовую инструкцию вместо кнопки.
	URL string `env:"WEBAPP_URL" envDefault:""`
}

type ModPortalConfig struct {
	Username        string `env:"FACTORIO_MOD_PORTAL_USER" envDefault:""`
	Token           string `env:"FACTORIO_MOD_PORTAL_TOKEN" envDefault:""`
	FactorioVersion string `env:"FACTORIO_VERSION" envDefault:"2.0"`
	ModsDir         string `env:"FACTORIO_MODS_DIR" envDefault:"/factorio/mods"`
	ModListFile     string `env:"FACTORIO_MOD_LIST_FILE" envDefault:"/factorio/mods/mod-list.json"`
}
