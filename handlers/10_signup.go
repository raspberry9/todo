package handlers

import (
	"net/http"

	"jsproj.com/koo/gosari/crypto"
	"jsproj.com/koo/gosari/utils"
	"jsproj.com/koo/server/auth/schema"
)

type qSignup struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type rSignup struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

const (
	signupOK                 = 0
	signupBadIDRequest       = -1010
	signupBadPasswordRequest = -1020
	signupServerError        = -1030
	signupIDDuplicated       = -1040 // 가입하려는 아이디가 이미 존재하는 경우
)

var signupErrors = map[int]string{
	defaultError: "Error occured during sign up.",

	signupBadIDRequest:       "Invalid email format.",
	signupBadPasswordRequest: "Invalid password format.",
	signupIDDuplicated:       "Email already exists.",
}

func signupError(res int) rSignup {
	msg, ok := signupErrors[res]
	if !ok {
		msg = signupErrors[defaultError]
	}
	return rSignup{res, msg}
}

// signupHandler 함수는 사용자의 가입을 처리한다.
func signupHandler(
	w http.ResponseWriter,
	r *http.Request,
	env *Environ,
) interface{} {
	var req qSignup
	Unmarshal(r, &req)

	// 아이디 형식 검사
	if !schema.IsValidIDFormat(req.ID) {
		return signupError(signupBadIDRequest)
	}

	// 비밀번호 형식 검사
	if !schema.IsValidPasswordFormat(req.Password) {
		return signupError(signupBadPasswordRequest)
	}

	// Salt를 앞쪽에 붙인 비밀번호를 만든다.
	salt := crypto.NewSalt(schema.SaltMaxSize)
	password := string(crypto.SecurePassword(salt, []byte(req.Password)))

	// config 파일 설정에 [activation] 섹션의 use가 true로 되어 있는 경우 사용자에게 이메일을
	// 보낸다. activation을 사용하지 않는 경우에는 바로 일반 유저로 가입 시킨다.
	var initialStatus int
	var activeKey string
	if env.Conf.IsUseActivation() {
		// Need email activation.
		initialStatus = schema.UserStatusDeactivated
		activeKey, _ = crypto.NewUUID()
	} else {
		// New account can use immediately.
		initialStatus = schema.UserStatusNormal
		activeKey = ""
	}

	// 사용자의 가입 정보를 수집하여 구조체로 만든다.
	user := &schema.User{
		ID:            req.ID,
		Info:          "{}",
		Status:        initialStatus,
		Password:      password,
		PasswordTmp:   "",
		Created:       utils.ServerTime(),
		ActivationKey: activeKey,
		Type:          schema.UserTypeNormal,
	}

	// 사용자를 데이터베이스에 저장한다.
	err := env.DB.Auth.Insert(user)
	if env.DB.IsDuplicated(err) {
		// 아이디가 중복됨
		return signupError(signupIDDuplicated)
	} else if err != nil {
		// 기타 데이터베이스 에러 발생
		return signupError(signupServerError)
	}

	if env.Conf.IsUseActivation() {
		// 인증 메일을 발송한다.
		if err := schema.SendMail(user, "activation_mail_title.tmpl", "activation_mail.tmpl"); err != nil {
			return signupError(signupServerError)
		}
	}

	return rSignup{signupOK, "success"}
}
