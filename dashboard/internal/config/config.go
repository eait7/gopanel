package config

import "os"

// Config holds all application configuration loaded from environment variables.
type Config struct {
	CaddyAPI            string
	PortainerURL        string
	PortainerExternalURL string
	FileBrowserURL      string
	Port                string
	Secret              string
	Username            string
	Password            string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		CaddyAPI:            getEnv("CADDY_API", "http://localhost:2019"),
		PortainerURL:        getEnv("PORTAINER_URL", "https://localhost:9443"),
		PortainerExternalURL: getEnv("PORTAINER_EXTERNAL_URL", "https://localhost:9443"),
		FileBrowserURL:      getEnv("FILEBROWSER_URL", "http://localhost:8090"),
		Port:                getEnv("GOPANEL_PORT", "9000"),
		Secret:              getEnv("GOPANEL_SECRET", "change-me-in-production"),
		Username:            getEnv("GOPANEL_USERNAME", "admin"),
		Password:            getEnv("GOPANEL_PASSWORD", "admin"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
