package wotop_jwt

import "github.com/a-aslani/wotop/model/apperror"

const (
	ErrUnauthorized                   apperror.ErrorType = "ER0001 unauthorized"
	ErrExpiredToken                   apperror.ErrorType = "ER0002 the token is expired"
	ErrTokenAlreadyRefreshed          apperror.ErrorType = "ER0003 the token is already refreshed"
	ErrRefreshTokenNotFoundInDatabase apperror.ErrorType = "ER0004 refresh token not found in database"
	ErrReadingJWTClaims               apperror.ErrorType = "ER0005 error reading jwt claims"
	ErrFetchingJWTClaims              apperror.ErrorType = "ER0006 error fetching claims"
	ErrParsingRefreshTokenWithClaims  apperror.ErrorType = "ER0007 could not parse refresh token with claims"
	ErrReadingRefreshTokenClaims      apperror.ErrorType = "ER0008 could not read refresh token claims"
)
