package handlers

import (
	"net/http"
	//"time"

	//log "github.com/Sirupsen/logrus"
	"jsproj.com/koo/gosari/crypto"
	"jsproj.com/koo/server/auth/schema"
)

type qLogin struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type rLogin struct {
	Res       int    `json:"res"`
	Msg       string `json:"msg"`
	Token     string `json:"token"`
	TempLogin bool   `json:"templogin"`
}

const (
	loginOK                 = 0
	loginBadIDRequest       = -1110
	loginBadPasswordRequest = -1120
	loginServerError        = -1130
	loginNoUserError        = -1140 // 사용자가 데이터베이스에 없음
	loginNotActivatedError  = -1150 // 아직 메일 인증을 완료하지 않음
	loginBlockUserError     = -1160 // 사용자가 블럭됨
	loginTokenIssueError    = -1199 // 토큰이 만료되었거나 아직 발급되지 않음
)

var loginErrors = map[int]string{
	defaultError: "Error occured during login.",

	loginBadIDRequest:      "Invalid email format.",
	loginNoUserError:       "Incorrect ID or Password.",
	loginBlockUserError:    "System has blocked your account. Please contact the support team for more information.",
	loginNotActivatedError: "Account not yet activated. Please check your email.",
	loginTokenIssueError:   "Error occured during issue token.",
}

func loginError(res int) rLogin {
	msg, ok := loginErrors[res]
	if !ok {
		msg = loginErrors[defaultError]
	}
	return rLogin{res, msg, "", false}
}

// loginHandler 함수는 사용자의 로그인을 처리한다.
func loginHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {

	var req qLogin
	var token string
	Unmarshal(r, &req)

	isTempLogin := false // 임시 로그인인 경우 바로 비밀번호를 변경해야 한다.

	//time.Sleep(3 * time.Second)

	// 아이디 형식 검사
	if !schema.IsValidIDFormat(req.ID) {
		return loginError(loginBadIDRequest)
	}

	// 비밀번호 형식 검사
	if !schema.IsValidPasswordFormat(req.Password) {
		// 비밀번호 형식이 틀렸지만 loginNoUserError로 발생시키는 이유는
		// 보안상의 이유로 ID가 존재하는지 여부를 확인할 수 없게 하기 위해서이다.
		return loginError(loginNoUserError)
	}

	// 데이터베이스에서 해당 유저를 불러온다.
	user, err := schema.LoadUserFromID(req.ID)
	if err != nil {
		return loginError(loginNoUserError)
	}

	// 아직 이메일 인증을 하지 않아서 로그인 불가능.
	if user.IsDeactivated() {
		return loginError(loginNotActivatedError)
	}

	// 정상 유저 상태가 아니다.(블럭 혹은 기타 사유로) 로그인을 금지 시킨다.
	if !user.IsNormal() {
		return loginError(loginBlockUserError)
	}

	// 비밀번호를 비교한다.
	salt := []byte(user.Password[:schema.SaltMaxSize])
	reqPass := string(crypto.SecurePassword(salt, []byte(req.Password)))
	if user.Password != reqPass {
		// 비밀번호가 틀리면 임시 비밀번호를 체크한다.
		if user.PasswordTmp != "" {
			salt = []byte(user.PasswordTmp[:schema.SaltMaxSize])
			reqPass = string(crypto.SecurePassword(salt, []byte(req.Password)))
			if user.PasswordTmp == reqPass {
				isTempLogin = true
			} else {
				return loginError(loginNoUserError)
			}
		} else {
			return loginError(loginNoUserError)
		}
	}

	// jwt 토큰을 발급하여 클라이언트에게 일려준다.
	token, err = schema.IssueToken(user)
	if err != nil {
		return loginError(loginTokenIssueError)
	}

	return rLogin{loginOK, "success", token, isTempLogin}
}
