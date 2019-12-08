package handlers

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"jsproj.com/koo/gosari/utils"
	"jsproj.com/koo/server/auth/schema"
)

// activationHandler 함수는 가입시 유저에게 보낸 메일의 링크를 클릭하면 활성화 처리한다.
// 따라서 로그인 전에 토큰 발급 없이 불릴 수 있어야 한다.
func activationHandler(
	w http.ResponseWriter,
	r *http.Request,
	env *Environ,
) interface{} {
	// url로부터  uuid를 얻어온다. 이 uuid가 activationKey가 된다.
	vars := mux.Vars(r)
	activationKey := vars["code"]

	// 얻어온 activationKey를 가진 유저를 데이터베이스에서 찾아온다.
	user, err := schema.LoadUserFromActivateionKey(activationKey)
	if err != nil || user == nil {
		body := fmt.Sprintf(
			"{\"res\":%d,\"msg\":\"%s\"}",
			actionPageNotFound,
			"activation key not found.")
		log.WithFields(log.Fields{
			"ip":   GetIP(r),
			"url":  r.URL.Path,
			"body": body,
		}).Debug("RES")

		http.NotFound(w, r)
		return nil
	}

	// 사용자의 상태가 비활성화 상태가 아니라면 404 not found를 띄워준다.
	if !user.IsDeactivated() {
		body := fmt.Sprintf(
			"{\"res\":%d,\"msg\":\"%s\"}",
			actionPageNotFound,
			"invalid user activation status.")
		log.WithFields(log.Fields{
			"ip":   GetIP(r),
			"url":  r.URL.Path,
			"body": body,
		}).Debug("RES")

		http.NotFound(w, r)
		return nil
	}

	// activationKey로 유저를 찾았으므로 해당 유저를 정상 가입 상태로 만든다.
	user.Status = schema.UserStatusNormal
	if _, err := env.DB.Auth.Update(user); err != nil {
		log.Panic(err)
	}

	// 유저에게 html 템플릿을 이용하여 성공 메시지를 보여준다.
	tmpl := env.Conf.TemplatePath("activation_ok.tmpl")
	utils.JoinTemplate(w, tmpl, user)
	w.Header().Set("Content-Type", "text/html")
	// 로그를 찍을 용도로만 사용되며 실제로 이 json이 사용자에게 보내지지는 않는다.
	body := fmt.Sprintf("{\"res\":%d,\"msg\":\"%s\"}", actionOK, "success")
	log.WithFields(log.Fields{
		"ip":   GetIP(r),
		"url":  r.URL.Path,
		"body": body,
	}).Debug("RES")

	return nil
}
