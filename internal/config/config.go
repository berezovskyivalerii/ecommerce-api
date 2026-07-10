package config

type Config struct {
	ServerPort    string
	DBURL         string
	AdminEmail    string
	AdminPassword string
}

func New(dbURL, adminEmail, adminPassword string) *Config {
	return &Config{
		ServerPort:    ":8080",
		DBURL:         dbURL,
		AdminEmail:    adminEmail,
		AdminPassword: adminPassword,
	}
}
