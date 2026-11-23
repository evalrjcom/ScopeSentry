package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestI18nNegotiatesSupportedLocale(t *testing.T) {
	router := gin.New()
	router.Use(I18nMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("language"))
	})

	for _, tt := range []struct {
		name, header, want string
	}{
		{"no preference", "", "zh-CN"},
		{"weighted English", "zh-CN;q=0.4,en-GB;q=0.9", "en-US"},
		{"fall back to supported preference", "fr-FR,en-US;q=0.5", "en-US"},
		{"malformed header", "en-US;q=invalid", "zh-CN"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Language", tt.header)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if got := rec.Header().Get("Content-Language"); got != tt.want {
				t.Errorf("Content-Language = %q, want %q", got, tt.want)
			}
			if got := rec.Body.String(); got != tt.want {
				t.Errorf("context language = %q, want %q", got, tt.want)
			}
			if got := rec.Header().Get("Vary"); got != "Accept-Language" {
				t.Errorf("Vary = %q, want Accept-Language", got)
			}
		})
	}
}
