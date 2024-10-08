package authpkg

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"time"

	aespkg "github.com/eden-quan/go-kratos-pkg/aes"
	threadpkg "github.com/eden-quan/go-kratos-pkg/thread"
	uuidpkg "github.com/eden-quan/go-kratos-pkg/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// Encryptor ...
type Encryptor interface {
	EncryptToString(plaintext, key string) (string, error)
	DecryptToString(ciphertext, key string) (string, error)
}

// TokenResponse ...
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
}

var _ AuthRepo = (*authRepo)(nil)

// AuthRepo ...
type AuthRepo interface {
	JWTSigningKeyFunc(ctx context.Context) jwt.Keyfunc
	JWTSigningMethod() jwt.SigningMethod
	JWTSigningClaims() jwt.Claims

	// SignToken 签证Token
	// @Param signKey 拼接在原来的signKey上
	SignToken(ctx context.Context, authClaims *Claims) (*TokenResponse, []*TokenItem, error)
	DecodeAccessToken(ctx context.Context, accessToken string) (*Claims, error)
	DecodeRefreshToken(ctx context.Context, refreshToken string) (*Claims, error)

	VerifyToken(ctx context.Context, jwtToken *jwt.Token) error
}

// Config ...
type Config struct {
	SigningMethod      *jwt.SigningMethodHMAC
	SignKey            string
	RefreshCrypto      Encryptor
	AuthCacheKeyPrefix *AuthCacheKeyPrefix
}

// authRepo ...
type authRepo struct {
	logHandler  *log.Helper
	config      *Config
	tokenManger TokenManger
}

// NewAuthRepo ...
func NewAuthRepo(redisCC redis.UniversalClient, logger log.Logger, config Config) (AuthRepo, error) {
	if config.SignKey == "" {
		return nil, fmt.Errorf("sign key is empty")
	}
	if config.SigningMethod == nil {
		config.SigningMethod = jwt.SigningMethodHS256
	}
	if config.RefreshCrypto == nil {
		config.RefreshCrypto = aespkg.NewCBCCipher()
	}
	config.AuthCacheKeyPrefix = CheckAuthCacheKeyPrefix(config.AuthCacheKeyPrefix)
	return &authRepo{
		logHandler:  log.NewHelper(log.With(logger, "module", "auth/repo")),
		config:      &config,
		tokenManger: NewTokenManger(redisCC, config.AuthCacheKeyPrefix),
	}, nil
}

// JWTSigningKeyFunc 密钥 jwt.Keyfunc
func (s *authRepo) JWTSigningKeyFunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.SignKey), nil
	}
}

// JWTSigningMethod 签名方法
func (s *authRepo) JWTSigningMethod() jwt.SigningMethod {
	return s.config.SigningMethod
}

// JWTSigningClaims 签名载体
func (s *authRepo) JWTSigningClaims() jwt.Claims {
	return &Claims{}
}

// SignToken ...
func (s *authRepo) SignToken(ctx context.Context, authClaims *Claims) (*TokenResponse, []*TokenItem, error) {
	// token
	if authClaims.ID == "" {
		authClaims.ID = uuidpkg.NewUUID()
	}
	tokenString, err := jwt.NewWithClaims(s.config.SigningMethod, authClaims).SignedString([]byte(s.config.SignKey))
	if err != nil {
		return nil, nil, fmt.Errorf("sign token failed: %w", err)
	}

	// refresh token
	refreshClaims := DefaultRefreshClaims(authClaims)
	refreshClaimsStr, err := refreshClaims.EncodeToString()
	if err != nil {
		return nil, nil, fmt.Errorf("encode refresh claims failed: %w", err)
	}
	refreshToken, err := s.config.RefreshCrypto.EncryptToString(refreshClaimsStr, s.config.SignKey)
	if err != nil {
		return nil, nil, fmt.Errorf("crypto refresh claims failed: %w", err)
	}

	// 存储
	var (
		userIdentifier = authClaims.Payload.UserIdentifier()
		tokenItems     = []*TokenItem{
			{
				TokenID:        authClaims.ID,
				RefreshTokenID: refreshClaims.ID,
				ExpiredAt:      authClaims.ExpiresAt.Time.Unix(),
				IsRefreshToken: false,
				Payload:        authClaims.Payload,
			},
			{
				TokenID:        authClaims.ID,
				RefreshTokenID: refreshClaims.ID,
				ExpiredAt:      refreshClaims.ExpiresAt.Time.Unix(),
				IsRefreshToken: true,
				Payload:        refreshClaims.Payload,
			},
		}
	)
	err = s.tokenManger.SaveTokens(ctx, userIdentifier, tokenItems)
	if err != nil {
		return nil, nil, err
	}

	// 登录限制
	threadpkg.GoSafe(func() {
		s.afterSignToken(ctx, authClaims)
	})

	res := &TokenResponse{
		AccessToken:  tokenString,
		RefreshToken: refreshToken,
	}
	return res, tokenItems, nil
}

// afterSignToken ...
func (s *authRepo) afterSignToken(ctx context.Context, authClaims *Claims) {
	checkErr := s.checkLoginLimit(ctx, authClaims)
	if checkErr != nil {
		s.logHandler.WithContext(ctx).Errorw("msg", "checkLoginLimit failed", "err", checkErr)
	}
	deleteErr := s.deleteExpireTokens(ctx, authClaims)
	if deleteErr != nil {
		s.logHandler.WithContext(ctx).Errorw("msg", "deleteExpireTokens failed", "err", deleteErr)
	}
}

// deleteExpireTokens 检查登录限制
func (s *authRepo) deleteExpireTokens(ctx context.Context, authClaims *Claims) error {
	var (
		userIdentifier = authClaims.Payload.UserIdentifier()
		nowUnix        = time.Now().Unix()
		expireList     []*TokenItem
	)

	allTokens, err := s.tokenManger.GetAllTokens(ctx, userIdentifier)
	if err != nil {
		return fmt.Errorf("GetAllTokens failed: %w", err)
	}
	for i := range allTokens {
		if allTokens[i].ExpiredAt > nowUnix {
			continue
		}
		expireList = append(expireList, allTokens[i])
	}

	// 删除过期
	if err := s.tokenManger.DeleteTokens(ctx, userIdentifier, expireList); err != nil {
		return fmt.Errorf("DeleteTokens failed: %w", err)
	}
	return nil
}

// checkLoginLimit 检查登录限制
func (s *authRepo) checkLoginLimit(ctx context.Context, authClaims *Claims) error {
	if authClaims.Payload.LoginLimit == LoginLimitEnum_UNLIMITED {
		return nil
	}
	userIdentifier := authClaims.Payload.UserIdentifier()
	allTokens, err := s.tokenManger.GetAllTokens(ctx, userIdentifier)
	if err != nil {
		return fmt.Errorf("GetAllTokens failed: %w", err)
	}

	var (
		blacklist []*TokenItem
		limitList []*TokenItem
	)
	for iKey := range allTokens {
		// 不检查刷新token
		if allTokens[iKey].IsRefreshToken {
			continue
		}
		// 跳过自己
		if allTokens[iKey].TokenID == authClaims.ID {
			continue
		}

		isLimit := false
		switch authClaims.Payload.LoginLimit {
		case LoginLimitEnum_ONLY_ONE:
			// 同一账户仅允许登录一次
			isLimit = true
		case LoginLimitEnum_PLATFORM_ONE:
			// 同一账户每个平台都可登录一次
			if authClaims.Payload.LoginPlatform == allTokens[iKey].Payload.LoginPlatform {
				isLimit = true
			}
		}
		if isLimit {
			blacklist = append(blacklist, allTokens[iKey])
			limitList = append(limitList, allTokens[iKey])
			if item, ok := allTokens[allTokens[iKey].RefreshTokenID]; ok {
				blacklist = append(blacklist, item)
			}
		}
	}

	// 添加黑名单
	if err := s.tokenManger.AddBlacklist(ctx, userIdentifier, blacklist); err != nil {
		return fmt.Errorf("AddBlacklist failed: %w", err)
	}
	// 添加登录限制
	if err := s.tokenManger.AddLoginLimit(ctx, limitList); err != nil {
		return fmt.Errorf("AddLoginLimit failed: %w", err)
	}
	return nil
}

// DecodeAccessToken ...
func (s *authRepo) DecodeAccessToken(ctx context.Context, accessToken string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(accessToken, claims, s.JWTSigningKeyFunc(ctx))
	if err != nil {
		err = fmt.Errorf("decrypt token failed: %w", err)
		return nil, err
	}
	// 验证有效性
	if err = claims.Valid(); err != nil {
		return nil, err
	}
	return claims, err
}

// DecodeRefreshToken ...
func (s *authRepo) DecodeRefreshToken(ctx context.Context, refreshToken string) (*Claims, error) {
	claimsStr, err := s.config.RefreshCrypto.DecryptToString(refreshToken, s.config.SignKey)
	if err != nil {
		err = fmt.Errorf("decrypt token failed: %w", err)
		return nil, err
	}
	claims := &Claims{}
	err = claims.DecodeString(claimsStr)
	if err != nil {
		err = fmt.Errorf("decode token claims failed: %w", err)
		return nil, err
	}
	// 验证有效性
	if err = claims.Valid(); err != nil {
		return nil, err
	}
	return claims, err
}

// VerifyToken 验证令牌
func (s *authRepo) VerifyToken(ctx context.Context, jwtToken *jwt.Token) error {
	authClaims, ok := jwtToken.Claims.(*Claims)
	if !ok {
		return ErrTokenInvalid()
	}

	// 黑名单
	isBlacklist, err := s.tokenManger.IsBlacklist(ctx, authClaims.ID)
	if err != nil {
		e := ErrInvalidClaims()
		e.Metadata = map[string]string{"err": err.Error()}
		return e
	}
	if isBlacklist {
		return ErrBlacklist()
	}

	// 白名单
	isExist, err := s.tokenManger.IsExistToken(ctx, authClaims.Payload.UserIdentifier(), authClaims.ID)
	if err != nil {
		e := ErrInvalidClaims()
		e.Metadata = map[string]string{"err": err.Error()}
		return e
	}
	if !isExist {
		return ErrWhitelist()
	}
	return nil
}
