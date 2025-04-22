package jwt

import (
	"context"
	"github.com/a-aslani/wotop/logger"
	"github.com/a-aslani/wotop/model/payload"
	"github.com/a-aslani/wotop/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// GinMiddleware provides middleware functionality for handling Token authentication
// and extracting access tokens from HTTP headers.
//
// Fields:
//   - log: An instance of the Logger interface for logging messages.
type GinMiddleware struct {
	log logger.Logger
}

// NewGinMiddleware creates a new instance of GinMiddleware.
//
// Parameters:
//   - log: An instance of the Logger interface for logging.
//
// Returns:
//   - A new GinMiddleware instance.
func NewGinMiddleware(log logger.Logger) GinMiddleware {
	return GinMiddleware{log: log}
}

// GetAccessTokenFromHeader extracts the access token from the "Authorization" header.
//
// The header must follow the format "Bearer <token>". If the header is missing,
// improperly formatted, or the token is empty, an error is returned.
//
// Parameters:
//   - c: The Gin context containing the HTTP request.
//
// Returns:
//   - token: The extracted access token.
//   - err: An error if the token cannot be extracted.
func (g GinMiddleware) GetAccessTokenFromHeader(c *gin.Context) (token string, err error) {

	if c.Request.Header["Authorization"] == nil || len(c.Request.Header["Authorization"]) == 0 {
		err = ErrUnauthorized
		return
	}

	authorization := strings.Split(c.Request.Header["Authorization"][0], " ")
	token = authorization[1]

	if authorization[0] != preTokenName {
		err = ErrUnauthorized
		return
	}

	if token == "" {
		err = ErrUnauthorized
		return
	}

	return
}

// Authentication is a middleware function for authenticating requests using Token.
//
// This middleware extracts the access token from the "Authorization" header,
// verifies the token, and sets the token claims in the Gin context. If the token
// is invalid or missing, the request is aborted with a 401 Unauthorized response.
//
// Parameters:
//   - jwt: An instance of the Token interface for verifying tokens.
//
// Returns:
//   - A Gin handler function for authentication.
func (g GinMiddleware) Authentication(jwt Token) gin.HandlerFunc {

	return func(c *gin.Context) {

		// Generate a unique trace ID for the request.
		traceID := util.GenerateID(16)
		ctx := logger.SetTraceID(context.Background(), traceID)

		// Extract the access token from the header.
		token, err := g.GetAccessTokenFromHeader(c)
		if err != nil {
			g.log.Error(ctx, err.Error())
			c.JSON(http.StatusUnauthorized, payload.NewErrorResponse(err, traceID))
			c.Abort()
			return
		}

		// Verify the token and extract claims.
		_, tokenClaims, err := jwt.VerifyToken(token)
		if err != nil {
			g.log.Error(ctx, err.Error())
			c.JSON(http.StatusUnauthorized, payload.NewErrorResponse(err, traceID))
			c.Abort()
			return
		}

		// Set token claims and user information in the Gin context.
		c.Set("TokenClaims", tokenClaims)
		c.Set("ID", tokenClaims.ID)
		c.Set("Role", tokenClaims.Role)

		// Proceed to the next middleware or handler.
		c.Next()
	}
}
