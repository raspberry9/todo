package handlers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"jsproj.com/koo/server/auth/schema"
)

type qConfig struct {
}

type rConfig struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

const (
	configOK               = 0
	configBadRequest       = -10
	configServerError      = -20
	configPermissionDenied = -30
	configBadConfgFile     = -40
)

// ensureAdminOrBlock 함수는 운영자인지 확인하고 만일 운영자가 아니라면 해당 유저를 블럭 시킨다.
func ensureAdminOrBlock(env *Environ) (bool, *rConfig) {
	if !env.Me.IsAdmin() {
		if err := env.Me.Block("config handler called with no permission"); err != nil {
			log.Panic(err)
		}
		return false, &rConfig{configPermissionDenied, "permission denied."}
	}
	return true, nil
}

// reloadConfigHandler 함수는 config 파일을 다시 읽는다.
func reloadConfigHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {
	var req qConfig
	Unmarshal(r, &req)

	// 운영자인지 확인하고 운영자가 아니리면 블럭 처리
	if ok, res := ensureAdminOrBlock(env); !ok {
		return res
	}

	// 설정 파일을 다시 읽는다.
	schema.LoadConfig()

	return rConfig{configOK, "success"}
}
