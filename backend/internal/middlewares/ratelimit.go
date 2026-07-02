package middlewares

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// limitadorJanela implementa um rate limiter simples por chave (IP), em
// memória, usando janela fixa. Leve, sem dependências externas — adequado a
// um servidor único em rede interna. Para mitigar força bruta no login.
type limitadorJanela struct {
	mu       sync.Mutex
	tentativas map[string]*registroTentativas
	limite   int
	janela   time.Duration
}

type registroTentativas struct {
	contagem  int
	inicio    time.Time
}

func novoLimitador(limite int, janela time.Duration) *limitadorJanela {
	l := &limitadorJanela{
		tentativas: make(map[string]*registroTentativas),
		limite:     limite,
		janela:     janela,
	}
	// Limpeza periódica de chaves expiradas para não vazar memória.
	go l.limparPeriodicamente()
	return l
}

func (l *limitadorJanela) limparPeriodicamente() {
	ticker := time.NewTicker(l.janela)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		agora := time.Now()
		for k, r := range l.tentativas {
			if agora.Sub(r.inicio) > l.janela {
				delete(l.tentativas, k)
			}
		}
		l.mu.Unlock()
	}
}

// permitir registra uma tentativa para a chave e devolve (permitido, esperar).
// Quando bloqueado, esperar indica quanto tempo falta para a janela reiniciar.
func (l *limitadorJanela) permitir(chave string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	agora := time.Now()
	r, ok := l.tentativas[chave]
	if !ok || agora.Sub(r.inicio) > l.janela {
		l.tentativas[chave] = &registroTentativas{contagem: 1, inicio: agora}
		return true, 0
	}

	if r.contagem >= l.limite {
		return false, l.janela - agora.Sub(r.inicio)
	}
	r.contagem++
	return true, 0
}

// RateLimitLogin limita tentativas de login por IP para mitigar força bruta.
// Padrão sugerido: até `limite` tentativas por `janela` (ex.: 10 por minuto).
func RateLimitLogin(limite int, janela time.Duration) gin.HandlerFunc {
	l := novoLimitador(limite, janela)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		permitido, esperar := l.permitir(ip)
		if !permitido {
			segundos := int(esperar.Seconds()) + 1
			c.Header("Retry-After", strconv.Itoa(segundos))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"erro": "Muitas tentativas de login. Aguarde alguns instantes e tente novamente.",
			})
			return
		}
		c.Next()
	}
}
