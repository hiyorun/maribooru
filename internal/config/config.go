package config

import (
	"fmt"
	"maribooru/internal/helpers"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
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
		Development   bool          `env:"DEVELOPMENT;default:false"`
		EnforceEmail  bool          `env:"ENFORCE_EMAIL;default:false"`
		AdminCreated  bool          `env:"ADMIN_CREATED;default:false"`
		TokenLifetime time.Duration `env:"TOKEN_LIFETIME;default:24h"`
	}

	Database struct {
		Username string `env:"DB_USERNAME;default:maridb"`
		Password string `env:"DB_PASSWORD;required"`
		Host     string `env:"DB_HOST;default:127.0.0.1"`
		Port     string `env:"DB_PORT;default:5432"`
		Name     string `env:"DB_NAME;default:maridb"`
		SSL      string `env:"DB_SSL;default:disable"`
	}

	HTTP struct {
		Host   string `env:"LISTEN_HOST;default:127.0.0.1"`
		Port   int    `env:"LISTEN_PORT;default:8080"`
		Domain string `env:"DOMAIN;default:http://localhost"`
	}

	JWT struct {
		Secret string         `env:"JWT_SECRET;required"`
		Config echojwt.Config `env:"-"`
	}

	AssetStorage struct {
		UseS3             bool   `env:"USE_S3;default:false"`
		S3Endpoint        string `env:"S3_ENDPOINT;required_if:USE_S3=true"`
		S3AccessKey       string `env:"S3_ACCESS_KEY;required_if:USE_S3=true"`
		S3SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY;required_if:USE_S3=true"`
		S3UseSSL          bool   `env:"S3_USE_SSL;default:false"`
	}
)

var log *zap.Logger

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	log = helpers.NewZapLogger(os.Getenv("DEVELOPMENT") == "true")
	if err != nil {
		log.Fatal("Error loading .env file: %v", zap.Error(err))
		return nil, err
	}

	c := &Config{}

	// Process struct fields recursively
	err = ProcessStruct(reflect.ValueOf(c).Elem(), "")
	if err != nil {
		log.Error("There was an error processing the config struct", zap.Error(err))
		return nil, err
	}

	c.JWT.Config = echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(helpers.JWTUser)
		},
		SigningKey: []byte(c.JWT.Secret),
	}

	return c, nil
}

func ProcessStruct(val reflect.Value, parentField string) error {
	missingRequired := []string{}

	t := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)
		fieldName := fieldType.Name
		if parentField != "" {
			fieldName = parentField + "." + fieldName
		}

		// If it's a struct, recursively process it
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Duration(0)) && fieldType.Tag.Get("env") != "-" {
			if err := ProcessStruct(field, fieldName); err != nil {
				return err
			}
			continue
		}

		tagValue := fieldType.Tag.Get("env")
		if tagValue == "" {
			continue
		}

		parts := strings.Split(tagValue, ";")
		envKey := parts[0]
		defaultValue := ""
		isRequired := false
		requiredIf := ""

		for _, part := range parts[1:] {
			switch {
			case strings.HasPrefix(part, "default:"):
				defaultValue = strings.TrimPrefix(part, "default:")
			case part == "required":
				isRequired = true
			case strings.HasPrefix(part, "required_if:"):
				requiredIf = strings.TrimPrefix(part, "required_if:")
			}
		}

		envValue := os.Getenv(envKey)
		if envValue == "" {
			envValue = defaultValue
		}

		if isRequired && envValue == "" {
			missingRequired = append(missingRequired, fieldName)
		}

		if requiredIf != "" {
			condition := strings.Split(requiredIf, "=")
			if len(condition) == 2 {
				conditionKey, conditionValue := condition[0], condition[1]
				if strings.EqualFold(os.Getenv(conditionKey), conditionValue) && envValue == "" {
					missingRequired = append(missingRequired, fmt.Sprintf("%s (required when %s=%s)", fieldName, conditionKey, conditionValue))
				}
			}
		}

		// Set field value based on its type
		switch field.Type() {
		case reflect.TypeOf(time.Duration(0)):
			if duration, err := time.ParseDuration(envValue); err == nil {
				field.Set(reflect.ValueOf(duration))
			} else {
				log.Fatal("Invalid duration", zap.String(envKey, err.Error()))
				return err
			}
		default:
			switch field.Kind() {
			case reflect.Int:
				if val, err := strconv.Atoi(envValue); err == nil {
					field.SetInt(int64(val))
				} else {
					log.Fatal("Invalid integer", zap.String(envKey, err.Error()))
					return err
				}
			case reflect.String:
				field.SetString(envValue)
			case reflect.Bool:
				field.SetBool(strings.ToLower(envValue) == "true")
			}
		}
	}

	if len(missingRequired) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingRequired, ", "))
	}

	return nil
}
