package authpkg

import (
	"context"

	aespkg "github.com/eden/go-kratos-pkg/aes"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/selector"
)

func ExampleServer() {
	var (
		redisCC   = &redis.Client{}
		signKey   = ""
		logger    = log.DefaultLogger
		whiteList = map[string]struct{}{}
	)
	authConfig := Config{
		SigningMethod:      jwt.SigningMethodHS256,
		SignKey:            signKey,
		RefreshCrypto:      aespkg.NewCBCCipher(),
		AuthCacheKeyPrefix: CheckAuthCacheKeyPrefix(nil),
	}
	repo, err := NewAuthRepo(redisCC, logger, authConfig)
	if err != nil {
		return
	}

	// ExampleWhiteListMatcher 路由白名单
	var ExampleWhiteListMatcher = func(whiteList map[string]struct{}) selector.MatchFunc {
		return func(ctx context.Context, operation string) bool {
			//if tr, ok := contextutil.MatchHTTPServerContext(ctx); ok {
			//	if _, ok := whiteList[tr.Request().URL.Path]; ok {
			//		return false
			//	}
			//}

			if _, ok := whiteList[operation]; ok {
				return false
			}
			return true
		}
	}

	_ = selector.Server(
		Server(
			repo.JWTSigningKeyFunc,
			WithSigningMethod(repo.JWTSigningMethod()),
			WithClaims(repo.JWTSigningClaims),
			WithTokenValidator(repo.VerifyToken),
		),
	).
		Match(ExampleWhiteListMatcher(whiteList)).
		Build()

	return
}
