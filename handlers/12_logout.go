package handlers

import (
	"net/http"
)

type qLogout struct {
}

type rLogout struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

const (
	logoutOK          = 0
	logoutBadRequest  = -1210
	logoutServerError = -1220
)

// logoutHandler 함수는 사용자를 로그아웃 처리한다.
func logoutHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {
	var req qLogout
	Unmarshal(r, &req)

	// TODO : JWT는 로그 아웃 시에 무엇을 해야 하나 확인 필요
	return rLogout{logoutOK, "success"}
}
