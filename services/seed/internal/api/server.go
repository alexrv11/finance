package api

import (
	"fmt"
	"net/http"

	"github.com/finance/seed/internal/ingest"
	"github.com/finance/seed/internal/store"
	"github.com/gin-gonic/gin"
)

// Server holds all API dependencies.
type Server struct {
	db        *store.DB
	scheduler *ingest.Scheduler
	router    *gin.Engine
}

func NewServer(db *store.DB, scheduler *ingest.Scheduler) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	s := &Server{db: db, scheduler: scheduler, router: r}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	v1 := s.router.Group("/api/v1")

	// Prices
	v1.GET("/prices", s.getPrices)
	v1.GET("/prices/latest", s.getLatestPrice)

	// Indicators
	v1.GET("/indicators", s.getIndicator)

	// Ingest control
	v1.POST("/ingest", s.triggerIngest)

	// Health
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (s *Server) Run(port string) error {
	return s.router.Run(fmt.Sprintf(":%s", port))
}
