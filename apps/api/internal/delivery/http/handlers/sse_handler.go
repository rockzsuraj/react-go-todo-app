package handlers

import (
	"fmt"
	"net/http"
	"time"

	"react-todos/apps/api/internal/delivery/http/middleware"
	"react-todos/apps/api/internal/domain/services"

	"github.com/google/uuid"
)

type SSEHandler struct {
	subscriber services.TodoEventSubscriber
}

func NewSSEHandler(subscriber services.TodoEventSubscriber) *SSEHandler {
	return &SSEHandler{subscriber: subscriber}
}

func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.UserIDFromContext(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch, unsubscribe := h.subscriber.SubscribeTodoChanges(userID)
	defer unsubscribe()

	// Initial success event
	_, _ = fmt.Fprintf(w, ": ok\n\n")
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()
		case event := <-ch:
			_, _ = fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
		}
	}
}
