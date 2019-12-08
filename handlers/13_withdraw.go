package handlers

import (
	"net/http"
)

// Withdraw request
type qWithdraw struct {
}

// Withdraw response
type rWithdraw struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

const (
	withdrawOK          = 0
	withdrawBadRequest  = -1310
	withdrawServerError = -1320
)

// withdrawHandler 함수는 사용자를 탈퇴 시킨다.
func withdrawHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {
	var req qWithdraw
	Unmarshal(r, &req)

	// 현재는 바로 DB에서 해당 사용자를 삭제한다.
	_, err := env.DB.Auth.Delete(env.Me)
	if err != nil {
		return rWithdraw{withdrawServerError, "database delete failed."}
	}

	return rWithdraw{withdrawOK, "success"}
}
