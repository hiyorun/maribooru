package middlewares

import (
	"fmt"
	"log"
	"maribooru/internal/account"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"maribooru/internal/permission"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gopkg.in/go-playground/assert.v1"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPermission(t *testing.T) {
	log.Println("Start test permission middleware")
	e := echo.New()

	cfg := &config.Config{
		JWT: config.JWT{
			Secret: "secret",
		},
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(account.User{}, permission.Permission{})

	log, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	mw := NewMiddleware(cfg, db, log)

	pass, err := helpers.PasswordHash("unittest")
	if err != nil {
		t.Fatal(err)
	}

	users := []*account.User{
		{
			ID:       uuid.New(),
			Name:     "moderator",
			Password: pass,
			Permission: permission.Permission{
				Permission: permission.Moderate | permission.Write | permission.Read,
			},
		},
		{
			ID:       uuid.New(),
			Name:     "approver",
			Password: pass,
			Permission: permission.Permission{
				Permission: permission.Approve | permission.Write | permission.Read,
			},
		},
		{
			ID:       uuid.New(),
			Name:     "approve-only",
			Password: pass,
			Permission: permission.Permission{
				Permission: permission.Approve,
			},
		},
		{
			ID:       uuid.New(),
			Name:     "rw",
			Password: pass,
			Permission: permission.Permission{
				Permission: permission.Read | permission.Write,
			},
		},
		{
			ID:       uuid.New(),
			Name:     "wo",
			Password: pass,
			Permission: permission.Permission{
				Permission: permission.Write,
			},
		},
		{
			ID:       uuid.New(),
			Name:     "ro",
			Password: pass,
			Permission: permission.Permission{
				Permission: permission.Read,
			},
		},
	}

	if err := db.Create(users).Error; err != nil {
		t.Fatal(err)
	}

	tests := []permission.Level{permission.Moderate, permission.Approve, permission.Read, permission.Write}

	for _, user := range users {
		jwt, err := helpers.GenerateJWT(user.ID, user.Name, cfg.JWT.Secret, time.Minute*2)
		if err != nil {
			t.Fatal(err)
		}

		for _, test := range tests {
			req := httptest.NewRequest(echo.GET, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

			mod := mw.PermissionMiddleware(test)
			h := mod(echo.HandlerFunc(func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			}))
			h(c)

			log.Debug("Checking permissions", zap.Any("user", user.Permission.Permission), zap.Any("against", test), zap.Any("bitwise", test&user.Permission.Permission))
			if test&user.Permission.Permission != 0 {
				log.Debug("Allowed")
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				log.Debug("Unauthorized")
				assert.Equal(t, http.StatusUnauthorized, rec.Code)
			}
		}
	}
}

func TestAdmin(t *testing.T) {
	log.Println("Start test permission middleware")
	e := echo.New()

	cfg := &config.Config{
		JWT: config.JWT{
			Secret: "secret",
		},
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(account.User{}, permission.Permission{}, account.Admin{})

	log, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	mw := NewMiddleware(cfg, db, log)

	pass, err := helpers.PasswordHash("unittest")
	if err != nil {
		t.Fatal(err)
	}

	users := []*account.User{
		{
			ID:       uuid.New(),
			Name:     "non-admin",
			Password: pass,
		},
		{
			ID:       uuid.New(),
			Name:     "admin",
			Password: pass,
		},
	}

	if err := db.Create(users).Error; err != nil {
		t.Fatal(err)
	}

	log.Debug("UUID is", zap.Any("UUID", users[0].ID))

	admin := account.Admin{
		UserID: users[0].ID,
	}

	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}

	jwt, err := helpers.GenerateJWT(users[0].ID, users[0].Name, cfg.JWT.Secret, time.Minute*2)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

	mod := mw.AdminMiddleware()
	h := mod(echo.HandlerFunc(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}))
	h(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	jwt, err = helpers.GenerateJWT(users[1].ID, users[1].Name, cfg.JWT.Secret, time.Minute*2)
	if err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest(echo.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

	h(c)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
