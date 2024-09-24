package api

import (
	"errors"
	"fmt"
	"maribooru/api/routes"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"net"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HTTPServer struct {
	db         *gorm.DB
	cfg        *config.Config
	httpServer *echo.Echo
	log        *zap.Logger
}

func NewHTTPServer(cfg *config.Config, db *gorm.DB, log *zap.Logger) HTTPServer {
	e := echo.New()
	e.Validator = helpers.NewValidator(validator.New())

	return HTTPServer{
		db:         db,
		cfg:        cfg,
		httpServer: e,
		log:        log,
	}
}

func (s *HTTPServer) testPort() (int, error) {
	port := s.cfg.HTTP.Port
	ln, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err == nil {
		_ = ln.Close()
		return port, nil
	}
	s.log.Warn("Port is in use, searching for next free port", zap.Any("port", port))

	for i := port; i <= 65535; i++ {
		nextLn, err := net.Listen("tcp", ":"+fmt.Sprint(i))
		if err == nil {
			_ = nextLn.Close()
			return i, nil
		}
	}
	return 0, errors.New("No free available ports")
}

func (s *HTTPServer) RunHTTPServer() {
	api := routes.InitVersionOne(s.httpServer, s.db, s.cfg, s.log)

	api.Settings()
	api.Users()
	api.Administrative()
	api.Heartbeat()

	openPort, err := s.testPort()
	if err != nil {
		s.log.Fatal("Failed to test port", zap.Error(err))
	}

	if err := s.httpServer.Start(fmt.Sprintf(":%d", openPort)); err != http.ErrServerClosed {
		s.log.Fatal("Failed to start HTTP Server", zap.Error(err))
	}
}
