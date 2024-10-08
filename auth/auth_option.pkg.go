package authpkg

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
)

// TokenValidateFunc 自定义验证
type TokenValidateFunc func(context.Context, *jwt.Token) error

// Option is jwt option.
type Option func(*options)

// Parser is a jwt parser
type options struct {
	signingMethod      jwt.SigningMethod
	claims             func() jwt.Claims
	tokenHeader        map[string]interface{}
	tokenValidatorFunc TokenValidateFunc
}

// WithSigningMethod with signing method option.
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

// WithClaims with customer claim
// If you use it in Server, f needs to return a new jwt.Claims object each time to avoid concurrent write problems
// If you use it in Client, f only needs to return a single object to provide performance
func WithClaims(f func() jwt.Claims) Option {
	return func(o *options) {
		o.claims = f
	}
}

// WithTokenHeader withe customer tokenHeader for client side
func WithTokenHeader(header map[string]interface{}) Option {
	return func(o *options) {
		o.tokenHeader = header
	}
}

// WithTokenValidator token验证
func WithTokenValidator(tokenValidator TokenValidateFunc) Option {
	return func(o *options) {
		o.tokenValidatorFunc = tokenValidator
	}
}
