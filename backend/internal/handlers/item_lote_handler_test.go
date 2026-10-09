package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/handlers"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCriarLoteHTTPComCamposPendentes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:lote-http?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Categoria{}, &models.Setor{}, &models.Servidor{}, &models.Item{}); err != nil {
		t.Fatal(err)
	}
	cat := &models.Categoria{Nome: "Monitor"}
	if err := db.Create(cat).Error; err != nil {
		t.Fatal(err)
	}
	itemRepo := repositories.NewItemRepository(db)
	svc := services.NewItemService(itemRepo, repositories.NewCategoriaRepository(db), repositories.NewSetorRepository(db), repositories.NewServidorRepository(db), nil, nil)
	h := handlers.NewItemHandler(svc, nil)
	r := gin.New()
	r.POST("/itens/lote", h.CriarLote)

	corpo, _ := json.Marshal(map[string]any{
		"descricao": "Monitor LG 24 polegadas", "categoria_id": cat.ID,
		"marca": "LG", "modelo": "24", "estado_conservacao": "novo", "unidades": 48,
	})
	req := httptest.NewRequest(http.MethodPost, "/itens/lote", bytes.NewReader(corpo))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("esperado 201, recebido %d: %s", res.Code, res.Body.String())
	}
	var resposta struct {
		LoteCodigo string `json:"lote_codigo"`
		IDs        []uint `json:"ids"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &resposta); err != nil {
		t.Fatal(err)
	}
	if resposta.LoteCodigo == "" || len(resposta.IDs) != 48 {
		t.Fatalf("lote HTTP incompleto: %+v", resposta)
	}
	var n int64
	if err := db.Model(&models.Item{}).Where("lote_codigo = ? AND numero_patrimonio IS NULL AND setor_id IS NULL", resposta.LoteCodigo).Count(&n).Error; err != nil || n != 48 {
		t.Fatalf("esperados 48 monitores pendentes, recebidos %d (%v)", n, err)
	}
}
