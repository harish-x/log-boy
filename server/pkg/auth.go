package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"slices"

	"server/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

var (
	keySetCache   jwk.Set      // cached key set
	lastFetch     time.Time    // last time the key set was fetched
	cacheMutex    sync.RWMutex // mutex for the cache
	lastFetchErr  error        // last error
	lastErrorTime time.Time
)

type UserClaims struct {
	UniqueName string `json:"unique_name"`
	Role       string `json:"scp"`
	Name       string `json:"name"`
}

func getKeySet(ctx context.Context) (jwk.Set, error) {
	cfg, err := config.SetupEnv()
	if err != nil {
		log.Fatal("Failed to load env variables", err)
	}

	url := fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/keys", cfg.DirectoryTenantID) // get public keys from microsoft
	set, err := jwk.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	return set, nil
}
func getCachedKeySet(ctx context.Context) (jwk.Set, error) {
	cacheMutex.RLock()

	if time.Since(lastFetch) < 5*time.Minute && keySetCache != nil { // cache for 5 minutes
		defer cacheMutex.RUnlock()
		return keySetCache, nil
	}
	if time.Since(lastErrorTime) < 1*time.Minute && lastFetchErr != nil { // cache for 1 minute if there is an error
		defer cacheMutex.RUnlock()
		return nil, lastFetchErr
	}
	cacheMutex.RUnlock()
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	set, err := getKeySet(ctx)
	if err != nil {
		lastFetchErr = err
		lastErrorTime = time.Now()
		return nil, err
	}

	keySetCache = set
	lastFetch = time.Now()
	lastFetchErr = nil
	return set, nil
}

func validateToken(tokenString string, aud string, iss string) (*UserClaims, error) {
	ctx := context.Background()

	keySet, err := getCachedKeySet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{}) // parse without validation
	if err != nil {
		return nil, fmt.Errorf("malformed token: %w", err)
	}

	kid, ok := token.Header["kid"].(string) // get key id from a token header
	if !ok {
		return nil, fmt.Errorf("missing kid in token header")
	}

	key, found := keySet.LookupKeyID(kid) // lookup key by key id
	if !found {
		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}

	var rawKey interface{}
	if err := key.Raw(&rawKey); err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}

	// validate token
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(iss),
		jwt.WithAudience(aud),
		jwt.WithLeeway(5*time.Minute),
	)

	verifiedToken, err := parser.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return rawKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)

	if !ok {
		if !hasRequiredScope(claims, "Data.Read") {
			return nil, fmt.Errorf("insufficient permissions")
		}
	}

	return &UserClaims{
		UniqueName: fmt.Sprintf("%v", claims["unique_name"]),
		Role:       fmt.Sprintf("%v", claims["scp"]),
		Name:       fmt.Sprintf("%v", claims["name"]),
	}, nil

}
func hasRequiredScope(claims jwt.MapClaims, requiredScope string) bool {
	scp, ok := claims["scp"].(string)
	if !ok {
		return false
	}
	scopes := strings.Split(scp, " ")
	return slices.Contains(scopes, requiredScope)
}

// handleTokenError returns the appropriate HTTP status code and error message
func handleTokenError(err error) (int, string) {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return fiber.StatusForbidden, "Session expired - please reauthenticate"

	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return fiber.StatusUnauthorized, "Token not yet valid"

	case strings.Contains(err.Error(), "invalid audience"):
		return fiber.StatusUnauthorized, "Invalid token audience"

	case strings.Contains(err.Error(), "invalid issuer"):
		return fiber.StatusUnauthorized, "Invalid token issuer"

	case strings.Contains(err.Error(), "signature"):
		return fiber.StatusUnauthorized, "Invalid token signature"

	case strings.Contains(err.Error(), "insufficient permissions"):
		return fiber.StatusForbidden, "Insufficient permissions"

	default:
		return fiber.StatusUnauthorized, "Invalid authentication token"
	}
}

// AuthMiddleware is a middleware that validates JWT tokens for REST API
func AuthMiddleware() fiber.Handler {

	cfg, err := config.SetupEnv()
	if err != nil {
		log.Fatal("Failed to load env variables", err)
	}

	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing Authorization header")
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid Authorization header format")
		}

		rawToken := authHeader[7:]
		rawToken = strings.TrimSpace(rawToken)
		if rawToken == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing token")
		}

		aud := fmt.Sprintf("api://%s", cfg.ApplicationClientID)
		iss := fmt.Sprintf("https://sts.windows.net/%s/", cfg.DirectoryTenantID)

		_, err := validateToken(rawToken, aud, iss)

		if err != nil {
			log.Printf("Token validation failed: %v", err)
			status, formatedErr := handleTokenError(err)
			return fiber.NewError(status, formatedErr)
		}

		return c.Next()
	}
}

// SSEAuthMiddleware is a middleware that validates JWT tokens for SSE
func SSEAuthMiddleware() fiber.Handler {

	cfg, err := config.SetupEnv()
	if err != nil {
		log.Fatal("Failed to load env variables", err)
	}
	return func(c *fiber.Ctx) error {
		token := c.Query("bearer")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "bearer is required"})
		}
		token = strings.TrimSpace(token)

		aud := fmt.Sprintf("api://%s", cfg.ApplicationClientID)
		iss := fmt.Sprintf("https://sts.windows.net/%s/", cfg.DirectoryTenantID)

		user, err := validateToken(token, aud, iss)

		if err != nil {
			log.Printf("Token validation failed: %v", err)
			status, formattedErr := handleTokenError(err)
			return fiber.NewError(status, formattedErr)
		}
		c.Locals("user", user)
		return c.Next()
	}
}
