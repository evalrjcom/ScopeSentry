package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

var supportedLanguages = language.NewMatcher([]language.Tag{
	language.SimplifiedChinese,
	language.AmericanEnglish,
})

// I18nMiddleware selects the best supported language for the response.
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Browsers send weighted language lists. Choose a supported locale
		// and use its canonical tag for translation and Content-Language.
		preferred, _, err := language.ParseAcceptLanguage(c.GetHeader("Accept-Language"))
		if err != nil {
			preferred = nil
		}
		_, index, _ := supportedLanguages.Match(preferred...)
		locale := []string{"zh-CN", "en-US"}[index]
		c.Set("language", locale)
		c.Writer.Header().Add("Vary", "Accept-Language")
		c.Header("Content-Language", locale)
		c.Next()
	}
}
