package config

import (
	"log"
	"maribooru/internal/helpers"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type (
	Config struct {
		AppConfig    AppConfig
		Database     Database
		HTTP         HTTP
		JWT          JWT
		AssetStorage AssetStorage
	}
	AppConfig struct {
		Development   bool
		EnforceEmail  bool
		AdminCreated  bool
		TokenLifetime time.Duration
	}
	Database struct {
		Username string
		Password string
		Host     string
		Port     string
		Name     string
		SSL      string
	}
	HTTP struct {
		Host   string
		Port   int
		Domain string
	}
	JWT struct {
		Secret []byte
		Config echojwt.Config
	}
	AssetStorage struct {
		Path string
	}
)

func LoadConfig() (*Config, error) {
	errEnv := godotenv.Load()

	if errEnv != nil {
		log.Fatal("Unable to load .env file", errEnv)
	}

	development, _ := configDefaults("DEVELOPMENT", "false")
	enforceEmail, _ := configDefaults("ENFORCE_EMAIL", "false")
	tokenStringLifetime, _ := configDefaults("TOKEN_LIFETIME", "24h")
	tokenLifetime, err := time.ParseDuration(tokenStringLifetime)
	if err != nil {
		log.Fatal("Token lifetime must be a duration")
	}

	dbUsername, _ := configDefaults("DB_USERNAME", "maridb")
	dbPassword, _ := configDefaults("DB_PASSWORD", "changeme")
	dbHost, _ := configDefaults("DB_HOST", "127.0.0.1")
	dbPort, _ := configDefaults("DB_PORT", "5432")
	dbName, _ := configDefaults("DB_NAME", "maridb")
	dbSSL, _ := configDefaults("DB_SSL", "disable")

	listenHost, _ := configDefaults("LISTEN_HOST", "127.0.0.1")
	listenPort, _ := configDefaults("LISTEN_PORT", "8080")
	intListenPort, err := strconv.Atoi(listenPort)
	if err != nil {
		log.Fatal("Port must be a number")
	}
	domain, _ := configDefaults("DOMAIN", "http://localhost")
	jwtSecret, _ := configDefaults("JWT_SECRET", "")
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(helpers.JWTUser)
		},
		SigningKey: []byte(jwtSecret),
	}
	storagePath, _ := configDefaults("ASSET_PATH", "./")

	var cfg Config = Config{
		AppConfig: AppConfig{
			Development:   development == "true",
			EnforceEmail:  enforceEmail == "true",
			TokenLifetime: tokenLifetime,
		},
		Database: Database{
			Username: dbUsername,
			Password: dbPassword,
			Host:     dbHost,
			Port:     dbPort,
			Name:     dbName,
			SSL:      dbSSL,
		},
		HTTP: HTTP{
			Host:   listenHost,
			Port:   intListenPort,
			Domain: domain,
		},
		JWT: JWT{
			Secret: []byte(jwtSecret),
			Config: config,
		},
		AssetStorage: AssetStorage{
			Path: storagePath,
		},
	}

	return &cfg, nil
}

func configDefaults(env, defaults string) (string, bool) {
	value, ok := os.LookupEnv(env)
	if !ok {
		log.Printf("%s is unset. Resorting to default value of %s", env, defaults)
		return defaults, ok
	}
	return value, ok
}
