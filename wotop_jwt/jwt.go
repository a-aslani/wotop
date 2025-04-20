package wotop_jwt

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
	"github.com/golang-jwt/jwt"
	"os"
	"strings"
	"time"
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

type JWTToken struct {
	algorithm             jwt.SigningMethod
	secretKey             string
	refreshTokenValidTime time.Duration
	accessTokenValidTime  time.Duration
	repo                  Repository
}

type Repository interface {
	StoreRefreshToken(ctx context.Context, sub, jti string) error
	StoreBlockedToken(ctx context.Context, sub, token string, expiresAt int64) error
	DeleteRefreshToken(ctx context.Context, jti string) error
	FindRefreshToken(ctx context.Context, jti string) (sub string, err error)
	FindAllRefreshTokens(ctx context.Context) ([]RefreshToken, error)
	FindAllBlockedTokens(ctx context.Context) ([]string, error)
}

type JWT interface {
	GenerateToken(ctx context.Context, userId string, role string, sub string, tenant string) (accessToken, refreshToken, csrfSecret string, expiresAt int64, err error)
	GenerateCentrifugoJWT(userId string, secretKey string) (string, error)
	RenewToken(ctx context.Context, oldAccessTokenString string, oldRefreshTokenString, oldCsrfSecret string) (newAccessToken, newRefreshToken, newCsrfSecret string, expiresAt int64, userId string, err error)
	DeleteToken(ctx context.Context, accessToken, refreshToken string) error
	VerifyToken(token string) (string, *Claims, error)
}

func NewHS256JWT(ctx context.Context, secretKey string, repo Repository, refreshTokenValidTime time.Duration, accessTokenValidTime time.Duration) (JWT, error) {

	jwtToken := &JWTToken{
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

func NewHS512JWT(ctx context.Context, secretKey string, repo Repository, refreshTokenValidTime time.Duration, accessTokenValidTime time.Duration) (JWT, error) {

	jwtToken := &JWTToken{
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

func NewRS256JWT(ctx context.Context, fileName string, repo Repository, refreshTokenValidTime time.Duration, accessTokenValidTime time.Duration) (JWT, error) {

	err := initRS256JWT(fileName)
	if err != nil {
		return nil, err
	}

	jwtToken := &JWTToken{
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

func initRS256JWT(fileName string) error {
	assetsDir := "assets"
	keysDir := "keys"
	path := fmt.Sprintf("%s/%s", assetsDir, keysDir)

	if _, err := os.Stat(fmt.Sprintf("./%s", assetsDir)); os.IsNotExist(err) {
		_ = os.Mkdir(fmt.Sprintf("./%s", assetsDir), 0755)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, 0755)
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s.rsa", path, fileName)); os.IsNotExist(err) {
		err = generateRSAKeys(path, fileName)
		if err != nil {
			return err
		}
	}

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

func generateRSAKeys(path string, fileName string) (err error) {
	// generate key
	privateKey, err := rsa.GenerateKey(cRand.Reader, 2048)
	if err != nil {
		return
	}
	publicKey := &privateKey.PublicKey

	// dump private key to file
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

	// dump public key to file
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

func (t *JWTToken) storeRefreshTokenToDatabase(ctx context.Context, sub, jti string) error {
	return t.repo.StoreRefreshToken(ctx, sub, jti)
}

func (t *JWTToken) storeBlockedTokenToDatabase(ctx context.Context, sub, token string, expiresAt int64) error {
	return t.repo.StoreBlockedToken(ctx, sub, token, expiresAt)
}

func (t *JWTToken) deleteRefreshTokenFromDatabase(ctx context.Context, jti string) error {
	return t.repo.DeleteRefreshToken(ctx, jti)
}

func (t *JWTToken) findRefreshTokenFromDatabase(ctx context.Context, jti string) (sub string, err error) {
	return t.repo.FindRefreshToken(ctx, jti)
}

func (t *JWTToken) findAllRefreshTokensFromDatabase(ctx context.Context) ([]RefreshToken, error) {
	return t.repo.FindAllRefreshTokens(ctx)
}

func (t *JWTToken) findAllBlockedTokensFromDatabase(ctx context.Context) ([]string, error) {
	return t.repo.FindAllBlockedTokens(ctx)
}

func (t *JWTToken) initCachedRefreshTokens(ctx context.Context) (err error) {

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

func (t *JWTToken) initCachedBlockedTokens(ctx context.Context) error {

	tokens, err := t.findAllBlockedTokensFromDatabase(ctx)
	if err != nil {
		return err
	}

	blockedTokens = tokens

	return nil
}

func (t *JWTToken) VerifyToken(authToken string) (string, *Claims, error) {

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

func (t *JWTToken) contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (t *JWTToken) verifyRefreshToken(refreshToken string) (*RefreshTokenClaims, error) {
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

func (t *JWTToken) storeRefreshToken(ctx context.Context, sub string) (jti string, err error) {
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

func (t *JWTToken) deleteRefreshToken(ctx context.Context, refreshToken string) (err error) {

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

func (t *JWTToken) DeleteToken(ctx context.Context, accessToken, refreshToken string) (err error) {

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

func (t *JWTToken) checkRefreshToken(jti string) bool {
	return refreshTokens[jti] != ""
}

func (t *JWTToken) generateCSRFSecret() (string, error) {
	return t.generateRandomString(32)
}

func (t *JWTToken) GenerateCentrifugoJWT(userId string, secretKey string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      userId,
		"channels": []string{"personal:broadcast"},
	}).SignedString([]byte(secretKey))
}

func (t *JWTToken) GenerateToken(ctx context.Context, userID string, role string, sub string, tenant string) (accessToken, refreshToken, csrfSecret string, expiresAt int64, err error) {

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

func (t *JWTToken) createAccessToken(userID string, role string, sub string, tenant string, csrfSecret string) (authTokenString string, authTokenExp int64, err error) {

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

func (t *JWTToken) RenewToken(ctx context.Context, oldAccessTokenString string, oldRefreshTokenString, oldCsrfSecret string) (newAuthTokenString, newRefreshTokenString, newCsrfSecret string, expiresAt int64, userId string, err error) {

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

func (t *JWTToken) parseToken(token *jwt.Token) (interface{}, error) {
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

func (t *JWTToken) updateRefreshTokenCsrf(oldRefreshTokenString string, newCsrfString string) (newRefreshTokenString string, err error) {
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

func (t *JWTToken) updateAccessToken(ctx context.Context, refreshTokenString string, oldAccessToken string) (newAccessToken, csrfSecret string, expiresAt int64, userId string, err error) {
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

func (t *JWTToken) sign(claims jwt.Claims) (string, error) {
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

func (t *JWTToken) updateRefreshTokenExp(ctx context.Context, oldRefreshTokenString string) (newRefreshTokenString string, err error) {
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

func (t *JWTToken) createRefreshToken(ctx context.Context, sub string, csrfString string) (refreshTokenString string, err error) {

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

func (t *JWTToken) grabUUID(authTokenString string) (string, error) {
	authToken, _ := jwt.ParseWithClaims(authTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return "", ErrFetchingJWTClaims
	})
	authTokenClaims, ok := authToken.Claims.(*Claims)
	if !ok {
		return "", ErrFetchingJWTClaims
	}

	return authTokenClaims.StandardClaims.Subject, nil
}

func (t *JWTToken) revokeRefreshToken(ctx context.Context, refreshTokenString string) error {
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

func (t *JWTToken) generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (t *JWTToken) generateRandomString(s int) (string, error) {
	b, err := t.generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
