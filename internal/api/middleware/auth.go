package middleware

import (
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required", nil)
			c.Abort()
			return
		}

		// 检查Bearer token格式
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid authorization header format", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Match the issuer's signing algorithm and require the expiry it sets.
		token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
			secret := config.GlobalConfig.JWT.Secret
			if secret == "" {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())

		if err != nil {
			logger.Error("Failed to parse token", zap.Error(err))
			response.Unauthorized(c, "Failed to parse token", nil)
			c.Abort()
			return
		}

		// 验证token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 将用户信息存储到上下文中
			c.Set("userID", claims["userID"])
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])
			c.Next()
		} else {
			response.Unauthorized(c, "Invalid token", nil)
			c.Abort()
			return
		}
	}
}
