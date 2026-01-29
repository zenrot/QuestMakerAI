package serv

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zenrot/QuestMakerAI/internal/api"
	"github.com/zenrot/QuestMakerAI/internal/llmClient/yandexAi"
)

type Server struct {
	router *mux.Router
}

func NewServer() *Server {
	s := &Server{
		router: mux.NewRouter(),
	}

	s.routes()
	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) routes() {
	s.router.HandleFunc(
		"/api/v1/questions",
		s.handleQuestions(),
	).Methods(http.MethodPost)
	f := http.FileServer(http.Dir("./static"))
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", f))
}

func (s *Server) handleQuestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(auth, "Api-Key ") {
			http.Error(w, "invalid Authorization format", http.StatusUnauthorized)
			return
		}

		apiKey := strings.TrimPrefix(auth, "Api-Key ")
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			http.Error(w, "empty api key", http.StatusUnauthorized)
			return
		}

		var req api.QuestionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		if req.Topic == "" || req.QuestionNumber <= 0 {
			http.Error(w, "invalid input parameters", http.StatusBadRequest)
			return
		}

		client := yandexAi.NewClient(apiKey)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		result, err := client.RequestTest(ctx, req.Topic, req.QuestionNumber)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
