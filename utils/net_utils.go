package utils

import (
	"github.com/gorilla/websocket"
	"strings"
)

/**
 * IsClosedOrGoingAwayCloseError check whether err is caused by closed connection or close-going-away connection
 */
func IsClosedOrGoingAwayCloseError(err error) bool {
	if err == nil {
		return false
	}
	closeErr, ok := err.(*websocket.CloseError)
	if ok {
		switch closeErr.Code {
		case websocket.CloseNormalClosure:
			return true
		case websocket.CloseGoingAway:
			return true
		default:
			return false
		}
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		// ugly error check because I don't know whether the error is normal closed connection read error
		return true
	}
	return false
}
