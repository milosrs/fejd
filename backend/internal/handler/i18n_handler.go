package handler

import (
	"net/http"

	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
)

type I18nHandler struct {
	translations *store.TranslationStore
}

func NewI18nHandler(translations *store.TranslationStore) *I18nHandler {
	return &I18nHandler{translations: translations}
}

// GetTranslations godoc
// @Summary      List UI labels for a locale
// @Description  Returns the app's UI labels as a flat key/value object for the given locale.
// @Tags         public
// @Produce      json
// @Param        locale path string true "Locale code (e.g. en, rs)"
// @Success      200 {object} Translations
// @Failure      500 {object} ErrorResponse
// @Router       /api/i18n/{locale} [get]
func (h *I18nHandler) GetTranslations(w http.ResponseWriter, r *http.Request) {
	locale := chi.URLParam(r, "locale")

	rows, err := h.translations.ListByLocale(r.Context(), locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get translations")
		return
	}

	out := make(Translations, len(rows))
	for _, tr := range rows {
		out[tr.Key] = tr.Value
	}

	writeJSON(w, http.StatusOK, out)
}
