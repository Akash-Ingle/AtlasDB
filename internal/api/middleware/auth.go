package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/atlasdb/atlasdb/internal/auth"
)

const userClaimsKey ctxKey = "user_claims"

func Auth(jwtMgr *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":{"code":"unauthorized","message":"Authorization header required"}}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":{"code":"unauthorized","message":"Invalid authorization format"}}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwtMgr.ValidateAccessToken(parts[1])
			if err != nil {
				http.Error(w, `{"error":{"code":"unauthorized","message":"Invalid or expired token"}}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(userClaimsKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
