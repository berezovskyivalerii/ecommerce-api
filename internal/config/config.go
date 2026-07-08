package config

type Config struct {
	ServerPort string
	DBURL      string
}

func New(dbURL string) *Config {
	return &Config{
		ServerPort: ":8080",
		DBURL:      dbURL,
	}
}
