package recaptcha

import "context"

//go:generate go run go.uber.org/mock/mockgen -destination mocks/recaptcha_mock.go -package mockrecaptcha ./ Recaptcha
type Recaptcha interface {
	SiteVerify(ctx context.Context, secret, token string) error
}
