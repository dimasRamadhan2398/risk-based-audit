package middleware

import (
	"strings"

	"kong-gateway/pkg/config"
	"kong-gateway/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthMiddleware creates an authentication middleware
type AuthMiddleware struct {
	secret     string
	skipPaths  map[string]bool
	tokenHeader string
	tokenPrefix string
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(cfg *config.AuthConfig) *AuthMiddleware {
	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return &AuthMiddleware{
		secret:     cfg.JWTSecret,
		skipPaths:  skipPaths,
		tokenHeader: cfg.TokenHeader,
		tokenPrefix: cfg.TokenPrefix,
	}
}

// Authenticate validates JWT tokens
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for certain paths
		if m.skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Check for Authorization header
		authHeader := c.GetHeader(m.tokenHeader)
		if authHeader == "" {
			response.Unauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		// Extract token from header
		tokenString := authHeader
		if m.tokenPrefix != "" && strings.HasPrefix(authHeader, m.tokenPrefix+" ") {
			tokenString = strings.TrimPrefix(authHeader, m.tokenPrefix+" ")
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.secret), nil
		})

		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		if !token.Valid {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("roles", claims.Roles)
		}

		c.Next()
	}
}

// RequireRoles checks if the user has any of the required roles
func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "No roles found")
			c.Abort()
			return
		}

		userRoleList, ok := userRoles.([]string)
		if !ok {
			response.Forbidden(c, "Invalid roles format")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		for _, requiredRole := range roles {
			for _, userRole := range userRoleList {
				if userRole == requiredRole {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// SkipAuth creates a middleware that skips authentication for specific paths
func (m *AuthMiddleware) SkipAuth(paths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, path := range paths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}
		// If not matching skip paths, run normal authentication
		m.Authenticate()(c)
	}
}
