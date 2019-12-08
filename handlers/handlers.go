// Package handlers implements handler functions for authentication http server.
package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"jsproj.com/koo/server/auth/schema"
)

// Environ 구조체는 각 유저의 패킷이 왔을 때 해당 요청을 보낸 유저 구조체와 설정파일 그리고 DB를
// 사용할 수 있도록 하는 문맥이다.
type Environ struct {
	Me   *schema.User      // 현재 처리 중인 사용자
	Conf *schema.Configure // 서버 설정 파일(schema.Config() 로도 접근 가능하다)
	DB   *schema.Databases // 데이터페이스 풀(schema.Database() 로도 접근 가능하다)
}

type actionFunc func(http.ResponseWriter, *http.Request, *Environ) interface{}

// GetIP 함수는 요청이 들어온 클라이언트의 IP 주소를 찾는다.
// Nginx등으로 포워딩 된경우 X-FORWARDED-FOR 헤더를 통해 실제 클라이언트의 IP 주소를 찾고
// 만일 해당 해더가 없는 경우에는 http.Request.RemoteAddr에서 포트번호를 떼고 IP 주소만 얻어온다.
func GetIP(r *http.Request) string {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return ipProxy
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// Unmarshal 함수는 요청이 들어온 body를 이용하여 원하는 구조체로 언마샬링 한다.
func Unmarshal(r *http.Request, m interface{}) error {
	body := reqLog(r)
	err := json.Unmarshal(body, m)
	if err != nil {
		return err
	}
	return nil
}

// Marshal 함수는 결과 구조체를 이용하여 json 형식의 문자열로 마샬링 한다.
func Marshal(r *http.Request, w http.ResponseWriter, m interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	byt, err := json.Marshal(m)
	if err != nil {
		return err
	}
	w.Write(byt)

	log.WithFields(log.Fields{
		"ip":   GetIP(r),
		"url":  r.URL.Path,
		"body": string(byt[:]),
	}).Debug("RES")

	return nil
}

type rAction struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

func (res rAction) mustSend(r *http.Request, w http.ResponseWriter) {
	err := Marshal(r, w, res)
	if err != nil {
		log.Panic(err)
	}
}

const (
	defaultError              = -9999
	actionOK                  = 200
	actionBadRequest          = -400
	actionUnauthorized        = -401
	actionPageNotFound        = -404
	actionInternalServerError = -500
)

func reqLog(r *http.Request) []byte {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	var lbody string
	if len(body) > 0 {
		lbody = string(body[:])
	} else {
		lbody = "{}"
	}

	log.WithFields(log.Fields{
		"ip":   GetIP(r),
		"url":  r.URL.Path,
		"body": string(lbody),
	}).Debug("REQ")

	return body
}

func processAction(
	isLoginRequired bool,
	f actionFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		if r.Method == "OPTIONS" {
			return
		}

		var err error
		var me *schema.User

		if isLoginRequired {
			me, err = schema.LoadUserFromRequest(r)
			if err != nil {
				reqLog(r)
				rAction{actionUnauthorized, "login required."}.mustSend(r, w)
				return
			}
			if me.IsBlocked() {
				reqLog(r)
				rAction{actionUnauthorized, "blocked."}.mustSend(r, w)
				return
			}
		}

		env := &Environ{
			Me:   me,
			Conf: schema.Config(),
			DB:   schema.Database()}

		res := f(w, r, env)
		if res == nil {
			return
		}
		err = Marshal(r, w, res)
		if err != nil {
			rAction{actionInternalServerError, "response marshal failed."}.mustSend(r, w)
			return
		}
	}
}

// Action function is a middleware of http handler function for non login required handlers.
func nonAction(f actionFunc) http.HandlerFunc {
	return processAction(false, f)
}

// Action function is a middleware of http handler function for login requred handlers.
func action(f actionFunc) http.HandlerFunc {
	return processAction(true, f)
}

// MustInit function is register Action and NonAction handler functions.
func MustInit() *mux.Router {
	r := mux.NewRouter()
	// nonAction 함수(로그인 하지 않은 상태에서 불리는 함수)
	r.HandleFunc("/activation/{code:[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}}", nonAction(activationHandler)).Methods("GET")
	r.HandleFunc("/signup", nonAction(signupHandler))

	// action 함수(로그인 된 후에만 부를 수 있는 함수)
	r.HandleFunc("/reloadconfig", action(reloadConfigHandler)).Methods("GET")
	r.HandleFunc("/logout", action(logoutHandler)).Methods("POST")
	r.HandleFunc("/withdraw", action(withdrawHandler)).Methods("POST")

	// 테스트용 함수
	r.HandleFunc("/hello", nonAction(helloHandler))

	// angularjs용 함수
	r.HandleFunc("/login", nonAction(loginHandler))
	r.HandleFunc("/todolist", action(todolistHandler))
	r.HandleFunc("/todosave", action(todoSaveHandler))
	r.HandleFunc("/todoremove", action(todoRemoveHandler))
	r.HandleFunc("/findpass", nonAction(findPassHandler))

	// 번역 js 파일
	//r.HandleFunc("/translate/{lang}", nonAction(translateHandler))

	return r
}
