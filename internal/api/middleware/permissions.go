package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MohamedElashri/snipo/internal/models"
)

// Permission levels
const (
	PermissionRead  = "read"
	PermissionWrite = "write"
	PermissionAdmin = "admin"
)

// GetTokenFromContext retrieves the API token from context
func GetTokenFromContext(ctx context.Context) *models.APIToken {
	if token, ok := ctx.Value(ContextKeyAPIToken).(*models.APIToken); ok {
		return token
	}
	return nil
}

// CheckPermission returns middleware that checks if the request has required permission level
func CheckPermission(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from context (set by RequireAuthWithTokenRepo middleware)
			token := GetTokenFromContext(r.Context())

			// If no token, this must be a session-based auth (full admin access)
			if token == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Check if token has required permission
			if !hasPermission(token.Permissions, required) {
				http.Error(w, `{"error":{"code":"INSUFFICIENT_PERMISSIONS","message":"Token does not have required permissions"}}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasPermission checks if the token's permission level is sufficient
func hasPermission(tokenPermission, required string) bool {
	// Admin has all permissions
	if tokenPermission == PermissionAdmin {
		return true
	}

	// Write has read + write permissions
	if tokenPermission == PermissionWrite {
		return required == PermissionRead || required == PermissionWrite
	}

	// Read only has read permission
	if tokenPermission == PermissionRead {
		return required == PermissionRead
	}

	return false
}

// RequireRead is a convenience middleware for read operations
func RequireRead(next http.Handler) http.Handler {
	return CheckPermission(PermissionRead)(next)
}

// RequireWrite is a convenience middleware for write operations
func RequireWrite(next http.Handler) http.Handler {
	return CheckPermission(PermissionWrite)(next)
}

// RequireAdmin is a convenience middleware for admin operations
func RequireAdmin(next http.Handler) http.Handler {
	return CheckPermission(PermissionAdmin)(next)
}

// PermissionByMethod returns middleware that checks permission based on HTTP method
// GET = read, POST/PUT/PATCH/DELETE = write
func PermissionByMethod(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToUpper(r.Method)
		
		var required string
		switch method {
		case "GET", "HEAD", "OPTIONS":
			required = PermissionRead
		case "POST", "PUT", "PATCH", "DELETE":
			required = PermissionWrite
		default:
			required = PermissionRead
		}

		CheckPermission(required)(next).ServeHTTP(w, r)
	})
}
