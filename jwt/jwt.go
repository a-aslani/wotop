package jwt

import (
	"context"
	"crypto/rand"
	cRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	RefreshTokenTableName = "refresh_token"
	BlockedTokenTableName = "blocked_token"
)

var (
	verifyKey     *rsa.PublicKey
	signKey       *rsa.PrivateKey
	refreshTokens map[string]string
	blockedTokens []string
	preTokenName  = "Bearer"
)

type Claims struct {
	ID     string `json:"id"`
	Csrf   string `json:"csrf"`
	Role   string `json:"role"`
	Tenant string `json:"tenant"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	Csrf string `json:"csrf"`
	jwt.StandardClaims
}

type RefreshToken struct {
	Subject string `json:"subject" bson:"subject"`
	JTI     string `json:"jti" bson:"jti"`
}

type token struct {
	algorithm             jwt.SigningMethod
	secretKey             string
	refreshTokenValidTime time.Duration
	accessTokenValidTime  time.Duration
	repo                  Repository
}

// Repository defines the interface for interacting with the token storage system.
// It provides methods for storing, retrieving, and deleting refresh tokens and blocked tokens.
type Repository interface {
	// StoreRefreshToken stores a refresh token in the database.
	// Parameters:
	// - ctx: The context for the operation.
	// - sub: The subject (user identifier) associated with the token.
	// - jti: The unique identifier for the token.
	// Returns:
	// - error: An error if the operation fails.
	StoreRefreshToken(ctx context.Context, sub, jti string) error

	// StoreBlockedToken stores a blocked token in the database.
	// Parameters:
	// - ctx: The context for the operation.
	// - sub: The subject (user identifier) associated with the token.
	// - token: The token string to be blocked.
	// - expiresAt: The expiration time of the blocked token (in Unix timestamp).
	// Returns:
	// - error: An error if the operation fails.
	StoreBlockedToken(ctx context.Context, sub, token string, expiresAt int64) error

	// DeleteRefreshToken deletes a refresh token from the database.
	// Parameters:
	// - ctx: The context for the operation.
	// - jti: The unique identifier of the token to be deleted.
	// Returns:
	// - error: An error if the operation fails.
	DeleteRefreshToken(ctx context.Context, jti string) error

	// FindRefreshToken retrieves a refresh token from the database.
	// Parameters:
	// - ctx: The context for the operation.
	// - jti: The unique identifier of the token to be retrieved.
	// Returns:
	// - sub: The subject (user identifier) associated with the token.
	// - error: An error if the operation fails.
	FindRefreshToken(ctx context.Context, jti string) (sub string, err error)

	// FindAllRefreshTokens retrieves all refresh tokens from the database.
	// Parameters:
	// - ctx: The context for the operation.
	// Returns:
	// - []RefreshToken: A list of all refresh tokens.
	// - error: An error if the operation fails.
	FindAllRefreshTokens(ctx context.Context) ([]RefreshToken, error)

	// FindAllBlockedTokens retrieves all blocked tokens from the database.
	// Parameters:
	// - ctx: The context for the operation.
	// Returns:
	// - []string: A list of all blocked token strings.
	// - error: An error if the operation fails.
	FindAllBlockedTokens(ctx context.Context) ([]string, error)
}

// Token defines the interface for managing JWT tokens.
// It provides methods for generating, renewing, deleting, and verifying tokens.
type Token interface {
	// GenerateToken generates a new access token, refresh token, and CSRF secret.
	// Parameters:
	// - ctx: The context for the operation.
	// - userId: The user ID for whom the token is generated.
	// - role: The role of the user.
	// - sub: The subject (user identifier) associated with the token.
	// - tenant: The tenant information for the user.
	// Returns:
	// - accessToken: The generated access token.
	// - refreshToken: The generated refresh token.
	// - csrfSecret: The generated CSRF secret.
	// - expiresAt: The expiration time of the access token (in Unix timestamp).
	// - error: An error if the operation fails.
	GenerateToken(ctx context.Context, userId string, role string, sub string, tenant string) (accessToken, refreshToken, csrfSecret string, expiresAt int64, err error)

	// GenerateCentrifugoJWT generates a JWT for Centrifugo.
	// Parameters:
	// - userId: The user ID for whom the token is generated.
	// - secretKey: The secret key used for signing the token.
	// Returns:
	// - string: The generated JWT.
	// - error: An error if the operation fails.
	GenerateCentrifugoJWT(userId string, secretKey string, capsObj map[string]interface{}) (string, error)

	// RenewToken renews an expired access token using a valid refresh token.
	// Parameters:
	// - ctx: The context for the operation.
	// - oldAccessTokenString: The expired access token string.
	// - oldRefreshTokenString: The refresh token string.
	// - oldCsrfSecret: The CSRF secret associated with the old tokens.
	// Returns:
	// - newAccessToken: The renewed access token.
	// - newRefreshToken: The renewed refresh token.
	// - newCsrfSecret: The new CSRF secret.
	// - expiresAt: The expiration time of the new access token (in Unix timestamp).
	// - userId: The user ID associated with the token.
	// - error: An error if the operation fails.
	RenewToken(ctx context.Context, oldAccessTokenString string, oldRefreshTokenString, oldCsrfSecret string) (newAccessToken, newRefreshToken, newCsrfSecret string, expiresAt int64, userId string, err error)

	// DeleteToken deletes an access token and its associated refresh token.
	// Parameters:
	// - ctx: The context for the operation.
	// - accessToken: The access token to be deleted.
	// - refreshToken: The refresh token to be deleted.
	// Returns:
	// - error: An error if the operation fails.
	DeleteToken(ctx context.Context, accessToken, refreshToken string) error

	// VerifyToken verifies the validity of an access token.
	// Parameters:
	// - token: The access token to be verified.
	// Returns:
	// - string: The token string if valid.
	// - *Claims: The claims extracted from the token.
	// - error: An error if the token is invalid or verification fails.
	VerifyToken(token string) (string, *Claims, error)
}

// NewHS256JWT creates a new JWT token instance using the HS256 signing method.
// Parameters:
// - ctx: The context for the operation.
// - secretKey: The secret key used for signing the token.
// - repo: The repository interface for token storage operations.
// - refreshTokenValidTime: The validity duration for refresh tokens.
// - accessTokenValidTime: The validity duration for access tokens.
// Returns:
// - Token: The created JWT token instance.
// - error: An error if the operation fails.
func NewHS256JWT(ctx context.Context, secretKey string, repo Repository, refreshTokenValidTime time.Duration, accessTokenValidTime time.Duration) (Token, error) {

	jwtToken := &token{
		algorithm:             jwt.SigningMethodHS256,
		secretKey:             secretKey,
		refreshTokenValidTime: refreshTokenValidTime,
		accessTokenValidTime:  accessTokenValidTime,
		repo:                  repo,
	}

	err := jwtToken.initCachedRefreshTokens(ctx)
	if err != nil {
		return nil, err
	}

	err = jwtToken.initCachedBlockedTokens(ctx)
	if err != nil {
		return nil, err
	}

	return jwtToken, nil
}

// NewHS512JWT creates a new JWT token instance using the HS512 signing method.
// Parameters:
// - ctx: The context for the operation.
// - secretKey: The secret key used for signing the token.
// - repo: The repository interface for token storage operations.
// - refreshTokenValidTime: The validity duration for refresh tokens.
// - accessTokenValidTime: The validity duration for access tokens.
// Returns:
// - Token: The created JWT token instance.
// - error: An error if the operation fails.
func NewHS512JWT(ctx context.Context, secretKey string, repo Repository, refreshTokenValidTime time.Duration, accessTokenValidTime time.Duration) (Token, error) {

	jwtToken := &token{
		algorithm:             jwt.SigningMethodHS512,
		secretKey:             secretKey,
		refreshTokenValidTime: refreshTokenValidTime,
		accessTokenValidTime:  accessTokenValidTime,
		repo:                  repo,
	}

	err := jwtToken.initCachedRefreshTokens(ctx)
	if err != nil {
		return nil, err
	}

	err = jwtToken.initCachedBlockedTokens(ctx)
	if err != nil {
		return nil, err
	}

	return jwtToken, nil
}

// NewRS256JWT creates a new JWT token instance using the RS256 signing method.
// Parameters:
// - ctx: The context for the operation.
// - fileName: The file name containing the RSA keys.
// - repo: The repository interface for token storage operations.
// - refreshTokenValidTime: The validity duration for refresh tokens.
// - accessTokenValidTime: The validity duration for access tokens.
// Returns:
// - Token: The created JWT token instance.
// - error: An error if the operation fails.
func NewRS256JWT(ctx context.Context, fileName string, repo Repository, refreshTokenValidTime time.Duration, accessTokenValidTime time.Duration) (Token, error) {

	err := initRS256JWT(fileName)
	if err != nil {
		return nil, err
	}

	jwtToken := &token{
		algorithm:             jwt.SigningMethodRS256,
		refreshTokenValidTime: refreshTokenValidTime,
		accessTokenValidTime:  accessTokenValidTime,
		repo:                  repo,
	}

	err = jwtToken.initCachedRefreshTokens(ctx)
	if err != nil {
		return nil, err
	}

	err = jwtToken.initCachedBlockedTokens(ctx)
	if err != nil {
		return nil, err
	}

	return jwtToken, nil
}

// initRS256JWT initializes the RSA keys for the RS256 signing method.
// It ensures the necessary directories and key files exist, and loads the keys into memory.
// Parameters:
// - fileName: The base name of the RSA key files (without extensions).
// Returns:
// - error: An error if the initialization fails.
func initRS256JWT(fileName string) error {
	assetsDir := "assets"
	keysDir := "keys"
	path := fmt.Sprintf("%s/%s", assetsDir, keysDir)

	// Ensure the assets directory exists
	if _, err := os.Stat(fmt.Sprintf("./%s", assetsDir)); os.IsNotExist(err) {
		_ = os.Mkdir(fmt.Sprintf("./%s", assetsDir), 0755)
	}

	// Ensure the keys directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, 0755)
	}

	// Generate RSA keys if they do not exist
	if _, err := os.Stat(fmt.Sprintf("%s/%s.rsa", path, fileName)); os.IsNotExist(err) {
		err = generateRSAKeys(path, fileName)
		if err != nil {
			return err
		}
	}

	// Load the private key
	privateKeyPath := fmt.Sprintf("%s/%s.rsa", path, fileName)
	publicKeyPath := fmt.Sprintf("%s/%s.rsa.pub", path, fileName)

	signBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}

	// Load the public key
	verifyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return err
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return err
	}

	return nil
}

// generateRSAKeys generates a new RSA key pair and saves them to files.
// Parameters:
// - path: The directory where the key files will be saved.
// - fileName: The base name of the RSA key files (without extensions).
// Returns:
// - error: An error if the key generation or file operations fail.
func generateRSAKeys(path string, fileName string) (err error) {
	// Generate a new RSA private key
	privateKey, err := rsa.GenerateKey(cRand.Reader, 2048)
	if err != nil {
		return
	}
	publicKey := &privateKey.PublicKey

	// Save the private key to a file
	var privateKeyBytes = x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	privatePem, err := os.Create(path + "/" + fileName + ".rsa")
	if err != nil {
		return
	}
	err = pem.Encode(privatePem, privateKeyBlock)
	if err != nil {
		return
	}

	// Save the public key to a file
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicPem, err := os.Create(path + "/" + fileName + ".rsa.pub")
	if err != nil {
		return
	}
	err = pem.Encode(publicPem, publicKeyBlock)
	if err != nil {
		return
	}

	return
}

// storeRefreshTokenToDatabase stores a refresh token in the database.
// Parameters:
// - ctx: The context for the operation.
// - sub: The subject (user identifier) associated with the token.
// - jti: The unique identifier for the token.
// Returns:
// - error: An error if the operation fails.
func (t *token) storeRefreshTokenToDatabase(ctx context.Context, sub, jti string) error {
	return t.repo.StoreRefreshToken(ctx, sub, jti)
}

// storeBlockedTokenToDatabase stores a blocked token in the database.
// Parameters:
// - ctx: The context for the operation.
// - sub: The subject (user identifier) associated with the token.
// - token: The token string to be blocked.
// - expiresAt: The expiration time of the blocked token (in Unix timestamp).
// Returns:
// - error: An error if the operation fails.
func (t *token) storeBlockedTokenToDatabase(ctx context.Context, sub, token string, expiresAt int64) error {
	return t.repo.StoreBlockedToken(ctx, sub, token, expiresAt)
}

// deleteRefreshTokenFromDatabase deletes a refresh token from the database.
// Parameters:
// - ctx: The context for the operation.
// - jti: The unique identifier of the token to be deleted.
// Returns:
// - error: An error if the operation fails.
func (t *token) deleteRefreshTokenFromDatabase(ctx context.Context, jti string) error {
	return t.repo.DeleteRefreshToken(ctx, jti)
}

// findRefreshTokenFromDatabase retrieves a refresh token from the database.
// Parameters:
// - ctx: The context for the operation.
// - jti: The unique identifier of the token to be retrieved.
// Returns:
// - sub: The subject (user identifier) associated with the token.
// - error: An error if the operation fails.
func (t *token) findRefreshTokenFromDatabase(ctx context.Context, jti string) (sub string, err error) {
	return t.repo.FindRefreshToken(ctx, jti)
}

// findAllRefreshTokensFromDatabase retrieves all refresh tokens from the database.
// Parameters:
// - ctx: The context for the operation.
// Returns:
// - []RefreshToken: A list of all refresh tokens.
// - error: An error if the operation fails.
func (t *token) findAllRefreshTokensFromDatabase(ctx context.Context) ([]RefreshToken, error) {
	return t.repo.FindAllRefreshTokens(ctx)
}

// findAllBlockedTokensFromDatabase retrieves all blocked tokens from the database.
// Parameters:
// - ctx: The context for the operation.
// Returns:
// - []string: A list of all blocked token strings.
// - error: An error if the operation fails.
func (t *token) findAllBlockedTokensFromDatabase(ctx context.Context) ([]string, error) {
	return t.repo.FindAllBlockedTokens(ctx)
}

// initCachedRefreshTokens initializes the cache for refresh tokens by loading them from the database.
// Parameters:
// - ctx: The context for the operation.
// Returns:
// - error: An error if the operation fails.
func (t *token) initCachedRefreshTokens(ctx context.Context) (err error) {

	refreshTokens = make(map[string]string)

	cachedRefreshTokens, err := t.findAllRefreshTokensFromDatabase(ctx)
	if err != nil {
		return
	}

	for _, token := range cachedRefreshTokens {
		refreshTokens[token.JTI] = token.Subject
	}

	return
}

// initCachedBlockedTokens initializes the cache for blocked tokens by loading them from the database.
// Parameters:
// - ctx: The context for the operation.
// Returns:
// - error: An error if the operation fails.
func (t *token) initCachedBlockedTokens(ctx context.Context) error {

	tokens, err := t.findAllBlockedTokensFromDatabase(ctx)
	if err != nil {
		return err
	}

	blockedTokens = tokens

	return nil
}

// VerifyToken verifies the validity of an access token.
// Parameters:
// - authToken: The access token to be verified.
// Returns:
// - string: The token string if valid.
// - *Claims: The claims extracted from the token.
// - error: An error if the token is invalid or verification fails.
func (t *token) VerifyToken(authToken string) (string, *Claims, error) {

	if len(strings.Split(authToken, " ")) > 1 {
		authToken = strings.Split(authToken, " ")[1]
	}

	token, err := jwt.ParseWithClaims(authToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.parseToken(token)
	})

	if err != nil {

		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				return authToken, nil, ErrExpiredToken
			}
		}

		return authToken, nil, ErrUnauthorized
	}

	if token.Valid {

		if t.contains(blockedTokens, authToken) {
			return authToken, nil, ErrUnauthorized
		}

		return authToken, token.Claims.(*Claims), nil
	} else {
		return authToken, nil, ErrUnauthorized
	}
}

// contains checks if a string exists in a slice of strings.
// Parameters:
// - s: The slice of strings to search.
// - e: The string to find.
// Returns:
// - bool: True if the string is found, false otherwise.
func (t *token) contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// verifyRefreshToken verifies the validity of a refresh token.
// Parameters:
// - refreshToken: The refresh token to be verified.
// Returns:
// - *RefreshTokenClaims: The claims extracted from the token.
// - error: An error if the token is invalid or verification fails.
func (t *token) verifyRefreshToken(refreshToken string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.parseToken(token)
	})

	if err != nil {

		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				return nil, ErrExpiredToken
			}
		}

		return nil, ErrUnauthorized
	}

	if token.Valid {
		return token.Claims.(*RefreshTokenClaims), nil
	} else {
		return nil, ErrUnauthorized
	}
}

// storeRefreshToken generates a unique identifier (JTI) for a refresh token, stores it in the database,
// and updates the in-memory cache of refresh tokens.
// Parameters:
// - ctx: The context for the operation.
// - sub: The subject (user identifier) associated with the token.
// Returns:
// - jti: The unique identifier for the refresh token.
// - error: An error if the operation fails.
func (t *token) storeRefreshToken(ctx context.Context, sub string) (jti string, err error) {
	jti, err = t.generateRandomString(32)
	if err != nil {
		return
	}

	for refreshTokens[jti] != "" {
		jti, err = t.generateRandomString(32)
		if err != nil {
			return
		}
	}

	err = t.storeRefreshTokenToDatabase(ctx, sub, jti)
	if err != nil {
		return
	}

	refreshTokens[jti] = sub

	return
}

// deleteRefreshToken deletes a refresh token from the database and removes it from the in-memory cache.
// Parameters:
// - ctx: The context for the operation.
// - refreshToken: The refresh token string to be deleted.
// Returns:
// - error: An error if the operation fails.
func (t *token) deleteRefreshToken(ctx context.Context, refreshToken string) (err error) {

	claims, err := t.verifyRefreshToken(refreshToken)
	if err != nil {
		return
	}

	sub, err := t.findRefreshTokenFromDatabase(ctx, claims.Id)
	if err != nil {
		return
	}

	token := RefreshToken{
		Subject: sub,
		JTI:     claims.Id,
	}

	if token.Subject != claims.Subject {
		return ErrRefreshTokenNotFoundInDatabase
	} else {

		err = t.deleteRefreshTokenFromDatabase(ctx, token.JTI)
		if err != nil {
			return
		}

		delete(refreshTokens, token.JTI)
	}

	return
}

// DeleteToken deletes an access token and its associated refresh token. If the access token is still valid,
// it is added to the blocked tokens list in the database and in-memory cache.
// Parameters:
// - ctx: The context for the operation.
// - accessToken: The access token to be deleted.
// - refreshToken: The refresh token to be deleted.
// Returns:
// - error: An error if the operation fails.
func (t *token) DeleteToken(ctx context.Context, accessToken, refreshToken string) (err error) {

	claims, err := t.verifyRefreshToken(refreshToken)
	if err != nil {
		return
	}

	sub, err := t.findRefreshTokenFromDatabase(ctx, claims.Id)
	if err != nil {
		return
	}

	token := RefreshToken{
		Subject: sub,
		JTI:     claims.Id,
	}

	if token.Subject != claims.Subject {
		return ErrRefreshTokenNotFoundInDatabase
	} else {
		err = t.deleteRefreshTokenFromDatabase(ctx, token.JTI)
		if err != nil {
			return
		}

		delete(refreshTokens, token.JTI)

		var accessClaims *Claims
		_, accessClaims, err = t.VerifyToken(accessToken)
		if err != nil {
			return
		}

		if accessClaims != nil && accessClaims.ExpiresAt != 0 && accessClaims.ExpiresAt > time.Now().Unix() {
			err = t.storeBlockedTokenToDatabase(ctx, token.Subject, accessToken, accessClaims.ExpiresAt)
			if err != nil {
				return
			}
			blockedTokens = append(blockedTokens, accessToken)
		}
	}

	return
}

// checkRefreshToken checks if a refresh token with the given JTI exists in the in-memory cache.
// Parameters:
// - jti: The unique identifier of the refresh token.
// Returns:
// - bool: True if the refresh token exists, false otherwise.
func (t *token) checkRefreshToken(jti string) bool {
	return refreshTokens[jti] != ""
}

// generateCSRFSecret generates a random CSRF secret string.
// Returns:
// - string: The generated CSRF secret.
// - error: An error if the operation fails.
func (t *token) generateCSRFSecret() (string, error) {
	return t.generateRandomString(32)
}

// GenerateCentrifugoJWT generates a JWT for Centrifugo with the specified user ID and secret key.
// Parameters:
// - userId: The user ID for whom the token is generated.
// - secretKey: The secret key used for signing the token.
// Returns:
// - string: The generated JWT.
// - error: An error if the operation fails.
func (t *token) GenerateCentrifugoJWT(userId string, secretKey string, capsObj map[string]interface{}) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      userId,
		"channels": []string{"personal:broadcast"},
		"caps":     []interface{}{capsObj},
	}).SignedString([]byte(secretKey))
}

// GenerateToken generates a new access token, refresh token, and CSRF secret.
// Parameters:
// - ctx: The context for the operation.
// - userID: The user ID for whom the token is generated.
// - role: The role of the user.
// - sub: The subject (user identifier) associated with the token.
// - tenant: The tenant information for the user.
// Returns:
// - accessToken: The generated access token.
// - refreshToken: The generated refresh token.
// - csrfSecret: The generated CSRF secret.
// - expiresAt: The expiration time of the access token (in Unix timestamp).
// - err: An error if the operation fails.
func (t *token) GenerateToken(ctx context.Context, userID string, role string, sub string, tenant string) (accessToken, refreshToken, csrfSecret string, expiresAt int64, err error) {

	// generate the csrf secret
	csrfSecret, err = t.generateCSRFSecret()
	if err != nil {
		return
	}

	// generate the refresh token
	refreshToken, err = t.createRefreshToken(ctx, sub, csrfSecret)

	// generate the auth token
	accessToken, expiresAt, err = t.createAccessToken(userID, role, sub, tenant, csrfSecret)
	if err != nil {
		return
	}

	return
}

// createAccessToken creates a new access token with the provided claims.
// Parameters:
// - userID: The user ID for whom the token is generated.
// - role: The role of the user.
// - sub: The subject (user identifier) associated with the token.
// - tenant: The tenant information for the user.
// - csrfSecret: The CSRF secret associated with the token.
// Returns:
// - authTokenString: The generated access token string.
// - authTokenExp: The expiration time of the access token (in Unix timestamp).
// - err: An error if the operation fails.
func (t *token) createAccessToken(userID string, role string, sub string, tenant string, csrfSecret string) (authTokenString string, authTokenExp int64, err error) {

	authTokenExp = time.Now().Add(t.accessTokenValidTime).Unix()
	authClaims := Claims{
		ID:     userID,
		Csrf:   csrfSecret,
		Role:   role,
		Tenant: tenant,
		StandardClaims: jwt.StandardClaims{
			Subject:   sub,
			ExpiresAt: authTokenExp,
		},
	}

	authTokenString, err = t.sign(authClaims)

	return
}

// RenewToken renews an expired access token using a valid refresh token and CSRF secret.
// Parameters:
// - ctx: The context for the operation.
// - oldAccessTokenString: The expired access token string.
// - oldRefreshTokenString: The refresh token string.
// - oldCsrfSecret: The CSRF secret associated with the old tokens.
// Returns:
// - newAuthTokenString: The renewed access token string.
// - newRefreshTokenString: The renewed refresh token string.
// - newCsrfSecret: The new CSRF secret.
// - expiresAt: The expiration time of the new access token (in Unix timestamp).
// - userId: The user ID associated with the token.
// - err: An error if the operation fails.
func (t *token) RenewToken(ctx context.Context, oldAccessTokenString string, oldRefreshTokenString, oldCsrfSecret string) (newAuthTokenString, newRefreshTokenString, newCsrfSecret string, expiresAt int64, userId string, err error) {

	if len(strings.Split(oldAccessTokenString, " ")) > 1 {
		oldAccessTokenString = strings.Split(oldAccessTokenString, " ")[1]
	}

	// first, check that a csrf token was provided
	if oldCsrfSecret == "" {
		fmt.Println("No CSRF token!")
		err = ErrUnauthorized
		return
	}

	// now, check that it matches what's in the auth token claims
	authToken, err := jwt.ParseWithClaims(oldAccessTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.parseToken(token)
	})

	authTokenClaims, ok := authToken.Claims.(*Claims)
	if !ok {
		return
	}

	if oldCsrfSecret != authTokenClaims.Csrf {
		fmt.Println("CSRF token doesn't match jwt!")
		err = ErrUnauthorized
		return
	}

	// next, check the auth token in a stateless manner
	if authToken.Valid {
		fmt.Println("Auth token is valid")
		// auth token has not expired
		// we need to return the csrf secret bc that's what the function calls for
		newCsrfSecret = authTokenClaims.Csrf

		// update the exp of refresh token string, but don't save to the db
		// we don't need to check if our refresh token is valid here
		// because we aren't renewing the auth token, the auth token is already valid
		newRefreshTokenString, err = t.updateRefreshTokenExp(ctx, oldRefreshTokenString)
		newAuthTokenString = oldAccessTokenString
		return
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		fmt.Println("Auth token is not valid")
		if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
			fmt.Println("Auth token is expired")
			// auth token is expired
			newAuthTokenString, newCsrfSecret, expiresAt, userId, err = t.updateAccessToken(ctx, oldRefreshTokenString, oldAccessTokenString)
			if err != nil {
				return
			}

			// update the exp of refresh token string
			newRefreshTokenString, err = t.updateRefreshTokenExp(ctx, oldRefreshTokenString)
			if err != nil {
				return
			}

			// update the csrf string of the refresh token
			newRefreshTokenString, err = t.updateRefreshTokenCsrf(newRefreshTokenString, newCsrfSecret)
			if err != nil {
				return
			}

			return
		} else {
			fmt.Println("Error in auth token")
			err = ErrUnauthorized
			return
		}
	} else {
		fmt.Println("Error in auth token")
		err = ErrUnauthorized
		return
	}

	// if we get here, there was some error validating the token
	err = ErrUnauthorized
	return
}

// parseToken parses a JWT token and validates its signing method.
// Parameters:
// - token: The JWT token to be parsed.
// Returns:
// - interface{}: The key used for signing the token.
// - error: An error if the token's signing method is invalid.
func (t *token) parseToken(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	var key interface{}

	switch t.algorithm {
	case jwt.SigningMethodRS256:
		key = verifyKey
	case jwt.SigningMethodHS256, jwt.SigningMethodHS512:
		key = []byte(t.secretKey)
	}

	return key, nil
}

// updateRefreshTokenCsrf updates the CSRF secret of a refresh token.
// Parameters:
// - oldRefreshTokenString: The old refresh token string.
// - newCsrfString: The new CSRF secret to be set.
// Returns:
// - newRefreshTokenString: The updated refresh token string.
// - err: An error if the operation fails.
func (t *token) updateRefreshTokenCsrf(oldRefreshTokenString string, newCsrfString string) (newRefreshTokenString string, err error) {
	refreshToken, err := jwt.ParseWithClaims(oldRefreshTokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.parseToken(token)
	})
	if err != nil {
		return
	}

	oldRefreshTokenClaims, ok := refreshToken.Claims.(*RefreshTokenClaims)
	if !ok {
		return
	}

	refreshClaims := RefreshTokenClaims{
		Csrf: newCsrfString,
		StandardClaims: jwt.StandardClaims{
			Id:        oldRefreshTokenClaims.StandardClaims.Id, // jti
			Subject:   oldRefreshTokenClaims.StandardClaims.Subject,
			ExpiresAt: oldRefreshTokenClaims.StandardClaims.ExpiresAt,
		},
	}

	newRefreshTokenString, err = t.sign(refreshClaims)
	return
}

// updateAccessToken updates an expired access token using a valid refresh token.
// Parameters:
// - ctx: The context for the operation.
// - refreshTokenString: The refresh token string.
// - oldAccessToken: The expired access token string.
// Returns:
// - newAccessToken: The updated access token string.
// - csrfSecret: The new CSRF secret.
// - expiresAt: The expiration time of the new access token (in Unix timestamp).
// - userId: The user ID associated with the token.
// - err: An error if the operation fails.
func (t *token) updateAccessToken(ctx context.Context, refreshTokenString string, oldAccessToken string) (newAccessToken, csrfSecret string, expiresAt int64, userId string, err error) {
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.parseToken(token)
	})
	if err != nil {
		return
	}

	refreshTokenClaims, ok := refreshToken.Claims.(*RefreshTokenClaims)
	if !ok {
		err = ErrReadingJWTClaims
		return
	}

	// check if the refresh token has been revoked
	if t.checkRefreshToken(refreshTokenClaims.StandardClaims.Id) {
		// the refresh token has not been revoked
		// has it expired?
		if refreshToken.Valid {
			// nope, the refresh token has not expired
			// issue a new auth token
			accessToken, _ := jwt.ParseWithClaims(oldAccessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return t.parseToken(token)
			})

			oldAuthTokenClaims, ok := accessToken.Claims.(*Claims)
			if !ok {
				err = ErrReadingJWTClaims
				return
			}

			// our policy is to regenerate the csrf secret for each new auth token
			csrfSecret, err = t.generateCSRFSecret()
			if err != nil {
				return
			}

			userId = oldAuthTokenClaims.ID

			newAccessToken, expiresAt, err = t.createAccessToken(oldAuthTokenClaims.ID, oldAuthTokenClaims.Role, oldAuthTokenClaims.StandardClaims.Subject, oldAuthTokenClaims.Tenant, csrfSecret)

			return
		} else {
			fmt.Println("Refresh token has expired!")
			// the refresh token has expired!
			// Revoke the token in our db and require the user to fmtin again
			err = t.DeleteToken(ctx, refreshTokenClaims.Subject, refreshTokenClaims.StandardClaims.Id)
			if err != nil {
				return
			}
			err = ErrUnauthorized
			return
		}
	} else {
		fmt.Println("Refresh token has been revoked!")
		// the refresh token has been revoked!
		err = ErrUnauthorized
		return
	}
}

// sign signs the provided claims and generates a JWT token string.
// Parameters:
// - claims: The claims to be signed.
// Returns:
// - string: The signed JWT token string.
// - error: An error if the signing operation fails.
func (t *token) sign(claims jwt.Claims) (string, error) {
	// create a signer
	token := jwt.NewWithClaims(t.algorithm, claims)

	var tokenString string
	var err error

	// generate the token string
	switch t.algorithm {
	case jwt.SigningMethodRS256:
		tokenString, err = token.SignedString(signKey)
		break
	case jwt.SigningMethodHS256, jwt.SigningMethodHS512:
		tokenString, err = token.SignedString([]byte(t.secretKey))
		break
	}

	return tokenString, err
}

// updateRefreshTokenExp updates the expiration time of a refresh token.
// Parameters:
// - ctx: The context for the operation.
// - oldRefreshTokenString: The old refresh token string.
// Returns:
// - newRefreshTokenString: The updated refresh token string.
// - err: An error if the operation fails.
func (t *token) updateRefreshTokenExp(ctx context.Context, oldRefreshTokenString string) (newRefreshTokenString string, err error) {
	refreshToken, err := jwt.ParseWithClaims(oldRefreshTokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.parseToken(token)
	})
	if err != nil {
		return
	}

	oldRefreshTokenClaims, ok := refreshToken.Claims.(*RefreshTokenClaims)
	if !ok {
		return
	}

	err = t.deleteRefreshToken(ctx, oldRefreshTokenString)
	if err != nil {
		return
	}

	refreshTokenExp := time.Now().Add(t.refreshTokenValidTime).Unix()

	refreshJti, err := t.storeRefreshToken(ctx, oldRefreshTokenClaims.StandardClaims.Subject)
	if err != nil {
		return
	}

	refreshClaims := RefreshTokenClaims{
		Csrf: oldRefreshTokenClaims.Csrf,
		StandardClaims: jwt.StandardClaims{
			Id:        refreshJti, // jti
			Subject:   oldRefreshTokenClaims.StandardClaims.Subject,
			ExpiresAt: refreshTokenExp,
		},
	}

	newRefreshTokenString, err = t.sign(refreshClaims)

	return
}

// createRefreshToken generates a new refresh token with the provided subject and CSRF string.
// Parameters:
// - ctx: The context for the operation.
// - sub: The subject (user identifier) associated with the token.
// - csrfString: The CSRF secret associated with the token.
// Returns:
// - refreshTokenString: The generated refresh token string.
// - err: An error if the operation fails.
func (t *token) createRefreshToken(ctx context.Context, sub string, csrfString string) (refreshTokenString string, err error) {

	refreshTokenExp := time.Now().Add(t.refreshTokenValidTime).Unix()

	refreshJti, err := t.storeRefreshToken(ctx, sub)
	if err != nil {
		return
	}

	refreshClaims := &RefreshTokenClaims{
		Csrf: csrfString,
		StandardClaims: jwt.StandardClaims{
			Id:        refreshJti, // jti
			Subject:   sub,
			ExpiresAt: refreshTokenExp,
		},
	}

	refreshTokenString, err = t.sign(refreshClaims)
	return
}

// grabUUID extracts the UUID (subject) from the provided access token string.
// Parameters:
// - authTokenString: The access token string to parse.
// Returns:
// - string: The extracted UUID (subject).
// - error: An error if the operation fails or the claims cannot be read.
func (t *token) grabUUID(authTokenString string) (string, error) {
	authToken, _ := jwt.ParseWithClaims(authTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return "", ErrFetchingJWTClaims
	})
	authTokenClaims, ok := authToken.Claims.(*Claims)
	if !ok {
		return "", ErrFetchingJWTClaims
	}

	return authTokenClaims.StandardClaims.Subject, nil
}

// revokeRefreshToken revokes a refresh token by deleting it from the database.
// Parameters:
// - ctx: The context for the operation.
// - refreshTokenString: The refresh token string to revoke.
// Returns:
// - error: An error if the operation fails or the token cannot be parsed.
func (t *token) revokeRefreshToken(ctx context.Context, refreshTokenString string) error {
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.parseToken(token)
	})
	if err != nil {
		return ErrParsingRefreshTokenWithClaims
	}

	refreshTokenClaims, ok := refreshToken.Claims.(*RefreshTokenClaims)
	if !ok {
		return ErrReadingRefreshTokenClaims
	}

	err = t.DeleteToken(ctx, refreshTokenClaims.Subject, refreshTokenClaims.StandardClaims.Id)
	if err != nil {
		return err
	}

	return nil
}

// generateRandomBytes generates a random byte slice of the specified length.
// Parameters:
// - n: The number of random bytes to generate.
// Returns:
// - []byte: The generated random byte slice.
// - error: An error if the random byte generation fails.
func (t *token) generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// generateRandomString generates a random string of the specified length.
// Parameters:
// - s: The length of the random string to generate.
// Returns:
// - string: The generated random string.
// - error: An error if the random byte generation fails.
func (t *token) generateRandomString(s int) (string, error) {
	b, err := t.generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
