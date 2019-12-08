package handlers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"jsproj.com/koo/gosari/crypto"
	"jsproj.com/koo/server/auth/schema"
)

type qFindPass struct {
	ID string `json:"id"`
}

type rFindPass struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

const (
	findPassOK           = 0
	findPassBadIDRequest = -1710
	findPassNoUserError  = -1720
	findPassServerError  = -1730
)

var findPassErrors = map[int]string{
	defaultError: "Error occured during find password.",

	findPassBadIDRequest: "Invalid email format.",
	findPassNoUserError:  "Incorrect ID.",
}

func findPassError(res int) rFindPass {
	msg, ok := findPassErrors[res]
	if !ok {
		msg = findPassErrors[defaultError]
	}
	return rFindPass{res, msg}
}

// findPassHandler 함수는 사용자의 비밀번호를 초기화 한다.
func findPassHandler(
	w http.ResponseWriter,
	r *http.Request,
	env *Environ,
) interface{} {
	var req qSignup
	Unmarshal(r, &req)

	// 아이디 형식 검사
	if !schema.IsValidIDFormat(req.ID) {
		return findPassError(findPassBadIDRequest)
	}

	// Salt를 앞쪽에 붙인 임시 비밀번호를 만든다.
	tmpPass, _ := crypto.NewUUID()
	tmpPass = tmpPass[0:8]
	salt := crypto.NewSalt(schema.SaltMaxSize)
	password := string(crypto.SecurePassword(salt, []byte(tmpPass)))

	user, err := schema.LoadUserFromID(req.ID)
	if err != nil {
		log.Debug(err)
		return findPassError(findPassNoUserError)
	}

	// 사용자의 임시 비밀번호를 업데이트 한다.
	user.PasswordTmp = password

	if _, err = env.DB.Auth.Update(user); err != nil {
		// 기타 데이터베이스 에러 발생
		log.Debug(err)
		return findPassError(findPassServerError)
	}

	// 보안상 좋지 않지만 이메일에 비밀번호 평문을 넣어야 하므로 임시로 tmpPass를 사용함. 이 이후에 user를 DB에 쓰는 일이 없도록 해야 함
	user.PasswordTmp = tmpPass
	if err := schema.SendMail(user, "findpass_mail_title.tmpl", "findpass_mail.tmpl"); err != nil {
		log.Debug(err)
		return findPassError(findPassServerError)
	}

	return rFindPass{findPassOK, "success"}
}
