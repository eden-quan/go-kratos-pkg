package websocketpkg

import (
	headerpkg "github.com/eden/go-kratos-pkg/header"
	"github.com/gorilla/websocket"
	"net/http"
)

// upgrade 升级http
var upgrade = &websocket.Upgrader{}

// DefaultUpgrade 默认升级
func DefaultUpgrade() *websocket.Upgrader {
	return upgrade
}

// UpgradeConn 升级链接
func UpgradeConn(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	cc, err := upgrade.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	headerpkg.SetIsWebsocket(r.Header)
	return cc, err
}
