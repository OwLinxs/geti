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
	catConsumo := &models.Categoria{Nome: "Cabo", Consumivel: true}
	if err := db.Create(catConsumo).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Item{Descricao: "Cabo de rede", CategoriaID: catConsumo.ID, Quantidade: 10, EstadoConservacao: models.EstadoConservacao("novo")}).Error; err != nil {
		t.Fatal(err)
	}
	itemRepo := repositories.NewItemRepository(db)
	svc := services.NewItemService(itemRepo, repositories.NewCategoriaRepository(db), repositories.NewSetorRepository(db), repositories.NewServidorRepository(db), nil, nil)
	h := handlers.NewItemHandler(svc, nil)
	r := gin.New()
	r.POST("/itens/lote", h.CriarLote)
	r.GET("/itens", h.Listar)

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
	for _, tc := range []struct {
		url   string
		total int
	}{
		{"/itens?tipo=patrimonio", 48},
		{"/itens?tipo=consumo", 1},
		{"/itens?tipo=patrimonio&sem_patrimonio=true&sem_setor=true", 48},
	} {
		res := httptest.NewRecorder()
		r.ServeHTTP(res, httptest.NewRequest(http.MethodGet, tc.url, nil))
		if res.Code != http.StatusOK {
			t.Fatalf("%s: status %d: %s", tc.url, res.Code, res.Body.String())
		}
		var lista struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &lista); err != nil {
			t.Fatal(err)
		}
		if lista.Total != tc.total {
			t.Fatalf("%s: esperado %d, recebido %d", tc.url, tc.total, lista.Total)
		}
	}
}
