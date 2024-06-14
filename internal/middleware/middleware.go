package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"

	"github.com/inodinwetrust10/mumbleBackend/utils"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from cookies
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Unauthorized - No token provided", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		secretKey := os.Getenv("JWT_SECRET")

		token, err := utils.ValidateJWT(tokenString, secretKey)
		if err != nil {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "id", claims["id"].(string))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
