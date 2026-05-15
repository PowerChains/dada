package main

import (
	context "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	cfg      Config
	db       *sql.DB
	cache    *redis.Client
	clients  map[*websocket.Conn]struct{}
	clientsM sync.Mutex
}

func NewServer(cfg Config, db *sql.DB, cache *redis.Client) *Server {
	return &Server{cfg: cfg, db: db, cache: cache, clients: map[*websocket.Conn]struct{}{}}
}

func (s *Server) SetupRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/best-price", s.corsMiddleware(s.bestPriceHandler))
	mux.HandleFunc("/api/v1/liquidity", s.corsMiddleware(s.liquidityHandler))
	mux.HandleFunc("/api/v1/simulate", s.corsMiddleware(s.simulateHandler))
	mux.HandleFunc("/api/v1/arbitrage", s.corsMiddleware(s.arbitrageHandler))
	mux.HandleFunc("/api/v1/trade-history", s.corsMiddleware(s.tradeHistoryHandler))
	mux.HandleFunc("/api/v1/ai", s.corsMiddleware(s.aiHandler))
	mux.HandleFunc("/ws/market", s.wsHandler)
	return mux
}

func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (s *Server) bestPriceHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	tokenIn := q.Get("tokenIn")
	tokenOut := q.Get("tokenOut")
	amount := parseFloat(q.Get("amount"), 1000)
	slippage := parseFloat(q.Get("slippageTolerance"), 0.5)

	resp, err := GetBestPrice(r.Context(), s.db, s.cache, tokenIn, tokenOut, amount, slippage)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) liquidityHandler(w http.ResponseWriter, r *http.Request) {
	pair := r.URL.Query().Get("pair")
	resp, err := AggregateLiquidity(r.Context(), s.db, pair)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) simulateHandler(w http.ResponseWriter, r *http.Request) {
	var req SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}
	resp, err := SimulateSwap(r.Context(), s.db, s.cache, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) arbitrageHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := DetectArbitrage(r.Context(), s.db, s.cache)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) tradeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	wallet := r.URL.Query().Get("wallet")
	if wallet == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "wallet is required"})
		return
	}
	trades, err := LoadTradeHistory(r.Context(), s.db, wallet)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, trades)
}

func (s *Server) aiHandler(w http.ResponseWriter, r *http.Request) {
	var req AIAssistantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ai payload"})
		return
	}
	answer, err := AskAiStudio(r.Context(), s.cfg, req.Message)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, AIAssistantResponse{Suggestion: answer})
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	s.clientsM.Lock()
	s.clients[conn] = struct{}{}
	s.clientsM.Unlock()

	for {
		if _, _, err := conn.NextReader(); err != nil {
			s.clientsM.Lock()
			delete(s.clients, conn)
			s.clientsM.Unlock()
			return
		}
	}
}

func (s *Server) broadcastArbitrage(items []ArbitrageOpportunity) {
	if len(items) == 0 {
		return
	}
	payload, err := json.Marshal(map[string]interface{}{"type": "arbitrage", "items": items})
	if err != nil {
		return
	}

	s.clientsM.Lock()
	defer s.clientsM.Unlock()
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseFloat(value string, fallback float64) float64 {
	if value == "" {
		return fallback
	}
	var parsed float64
	_, err := fmt.Sscan(value, &parsed)
	if err != nil {
		return fallback
	}
	return parsed
}
