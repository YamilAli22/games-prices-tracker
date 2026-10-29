// internal/debug/handler.go — paquete nuevo
package debug

import (
	"net/http"
	"github.com/YamilAli22/content_tracker/internal/worker"
)

type Handler struct {
	Worker *worker.Worker
}

func (h *Handler) HandleCheckPrices(w http.ResponseWriter, r *http.Request) {
	if err := h.Worker.CheckAllPrices(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Check completado, revisá los logs"))
}
