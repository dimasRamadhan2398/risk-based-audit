package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims mirrors the token auth-service issues. Keep the json tags in step with
// auth-service/pkg/middleware/auth.go or roles silently arrive empty.
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	// EmployeeID / EmployeeCode link the login to master-data employees; set by
	// auth-service at login, empty for unlinked accounts and older tokens.
	EmployeeID   string `json:"employee_id,omitempty"`
	EmployeeCode string `json:"employee_code,omitempty"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates the JWT auth-service issued.
//
// risk-service shipped without this: Kong has no jwt/key-auth plugin, so the
// whole risk register and RCM — including DELETE /api/v1/risks/:id — was
// reachable by anyone who could reach the gateway.
type AuthMiddleware struct {
	secret string
}

// NewAuthMiddleware fails rather than returning a middleware that cannot
// authenticate anything. An empty secret would make every token signed with ""
// valid, which is worse than an obvious startup crash.
func NewAuthMiddleware(secret string) (*AuthMiddleware, error) {
	if err := ValidateJWTSecret(secret); err != nil {
		return nil, err
	}
	return &AuthMiddleware{secret: secret}, nil
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error":   gin.H{"code": "UNAUTHORIZED", "message": message},
	})
}

func forbidden(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"success": false,
		"error":   gin.H{"code": "FORBIDDEN", "message": message},
	})
}

// Authenticate validates the bearer token and puts the caller on the context.
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := bearerToken(c)
		if tokenString == "" {
			unauthorized(c, "Missing authorization header")
			return
		}

		// The signing method is pinned: without WithValidMethods a token is
		// only as trustworthy as the weakest algorithm the library accepts.
		// ExpirationRequired rejects tokens that simply omit "exp".
		token, err := jwt.ParseWithClaims(
			tokenString,
			&Claims{},
			func(*jwt.Token) (interface{}, error) { return []byte(m.secret), nil },
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !token.Valid {
			unauthorized(c, "Invalid or expired token")
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			unauthorized(c, "Invalid token claims")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("employee_id", claims.EmployeeID)
		c.Set("employee_code", claims.EmployeeCode)
		c.Next()
	}
}

// bearerToken reads the token from the Authorization header only. Query-string
// and cookie tokens are refused: the request URI is written to access logs, and
// cookies would make the API open to cross-site requests.
func bearerToken(c *gin.Context) string {
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if parts := strings.Fields(authHeader); len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	return ""
}

// RequireRoles rejects callers that hold none of the given roles. Authenticate
// must run first.
func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			forbidden(c, "No roles found")
			return
		}

		userRoleList, ok := userRoles.([]string)
		if !ok {
			forbidden(c, "Invalid roles format")
			return
		}

		for _, required := range roles {
			for _, held := range userRoleList {
				if held == required {
					c.Next()
					return
				}
			}
		}

		forbidden(c, "Insufficient permissions")
	}
}
