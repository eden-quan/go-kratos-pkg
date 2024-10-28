package contextpkg

import (
	headerpkg "gitlab.lainuoniao.cn/rhinobird/backend/go-kratos-pkg.git/header"
)

// TrustedPlatform 信任的平台
var (
	defaultTrustedPlatform = headerpkg.RemoteAddr
)

// SetTrustedPlatform 设置信任的平台
func SetTrustedPlatform(platformHeader string) {
	defaultTrustedPlatform = platformHeader
}
