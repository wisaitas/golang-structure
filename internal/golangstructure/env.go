package golangstructure

var ENV struct {
	Port     string `env:"PORT" envDefault:"8080"`
	Postgres struct {
		Host     string `env:"HOST" envDefault:"localhost"`
		Port     string `env:"PORT" envDefault:"5432"`
		User     string `env:"USER" envDefault:"postgres"`
		Password string `env:"PASSWORD" envDefault:"postgres"`
		DBName   string `env:"DB_NAME" envDefault:"postgres"`
		SSLMode  string `env:"SSL_MODE" envDefault:"disable"`
	} `envPrefix:"POSTGRES_"`
}
