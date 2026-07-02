package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler expõe verificações de saúde da aplicação.
type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health responde se a API e o banco de dados estão operacionais.
func (h *HealthHandler) Health(c *gin.Context) {
	dbStatus := "ok"
	httpStatus := http.StatusOK

	sqlDB, err := h.db.DB()
	if err != nil {
		dbStatus = "erro"
		httpStatus = http.StatusServiceUnavailable
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "indisponível"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"servico":   "SIGE-TI API",
		"status":    statusText(httpStatus),
		"banco":     dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func statusText(code int) string {
	if code == http.StatusOK {
		return "operacional"
	}
	return "degradado"
}
