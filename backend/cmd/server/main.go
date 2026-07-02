package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/container"
	"github.com/pmfb/sige-ti/internal/database"
	"github.com/pmfb/sige-ti/internal/router"
	"github.com/pmfb/sige-ti/internal/seed"
)

func main() {
	// Flags de linha de comando.
	var (
		rodarSeed = flag.Bool("seed", false, "popula o banco com dados iniciais e encerra")
	)
	flag.Parse()

	// Carrega configuração a partir do ambiente.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[config] erro ao carregar configuração: %v", err)
	}

	// Conecta ao banco de dados (dialeto definido por configuração).
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[database] %v", err)
	}

	// Aplica as migrações (cria/atualiza o schema).
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("[database] %v", err)
	}
	log.Println("[database] migrações aplicadas")

	// Monta o container de dependências.
	ct := container.New(cfg, db)

	// Modo seed: popula e encerra.
	if *rodarSeed {
		if err := seed.Executar(ct); err != nil {
			log.Fatalf("[seed] %v", err)
		}
		log.Println("[seed] finalizado")
		return
	}

	// Seed automático na inicialização quando SEED_ON_START=true.
	if os.Getenv("SEED_ON_START") == "true" {
		if err := seed.Executar(ct); err != nil {
			log.Printf("[seed] aviso: %v", err)
		}
	}

	engine := router.Setup(cfg, ct)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Inicia o servidor em goroutine para permitir shutdown gracioso.
	go func() {
		log.Printf("[server] SIGE-TI API ouvindo em http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[server] erro ao iniciar: %v", err)
		}
	}()

	// Aguarda sinal de interrupção para encerrar com segurança.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[server] encerrando...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[server] shutdown forçado: %v", err)
	}
	log.Println("[server] encerrado com sucesso")
}
