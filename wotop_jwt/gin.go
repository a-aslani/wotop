package wotop_jwt

import (
	"context"
	"github.com/a-aslani/wotop/model/payload"
	"github.com/a-aslani/wotop/wotop_logger"
	"github.com/a-aslani/wotop/wotop_util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type GinMiddleware struct {
	log wotop_logger.Logger
}

func NewGinMiddleware(log wotop_logger.Logger) GinMiddleware {
	return GinMiddleware{log: log}
}

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

func (g GinMiddleware) Authentication(jwt JWT) gin.HandlerFunc {

	return func(c *gin.Context) {

		traceID := wotop_util.GenerateID(16)
		ctx := wotop_logger.SetTraceID(context.Background(), traceID)

		token, err := g.GetAccessTokenFromHeader(c)
		if err != nil {
			g.log.Error(ctx, err.Error())
			c.JSON(http.StatusUnauthorized, payload.NewErrorResponse(err, traceID))
			c.Abort()
			return
		}

		_, tokenClaims, err := jwt.VerifyToken(token)
		if err != nil {
			g.log.Error(ctx, err.Error())
			c.JSON(http.StatusUnauthorized, payload.NewErrorResponse(err, traceID))
			c.Abort()
			return
		}

		c.Set("TokenClaims", tokenClaims)
		c.Set("ID", tokenClaims.ID)
		c.Set("Role", tokenClaims.Role)

		c.Next()
	}
}
