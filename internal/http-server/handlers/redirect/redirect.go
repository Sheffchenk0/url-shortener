package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Request struct {
	Alias string `json:"alias"`
}

type Response struct {
	resp.Response
}

type URLGetter interface {
	GetUrl(alias string) (string, error) 
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("alias is empty"))

			return
		}

		url, err := urlGetter.GetUrl(alias)
		
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("URL not found", slog.String("alias", alias))
		
			render.JSON(w, r, resp.Error("URL not found"))
			
			return
		}

		if err != nil {
			log.Error("failed to get URL", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to get URL"))

			return
		}

		log.Info("got url", slog.String("url", url))

		// redirect to found url
		http.Redirect(w, r, url, http.StatusFound)
	}
}
