package schema

import (
	"bytes"
	"encoding/gob"
	"net/mail"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"gopkg.in/gorp.v1"
	"jsproj.com/koo/gosari/utils"
)

// User 객체는 사용자 스키마 객체이다. 여기에서 정의된 형태로 데이터베이스 테이블이 작성된다.
// 따라서 이 코드를 바꾸는 경우 반드시 DB 마이그레이션이 필요하다.
type User struct {
	UID           int64  `db:"uid" json:"uid"`       // User's unique id.
	ID            string `db:"id" json:"id"`         // User's email id.
	Info          string `db:"info" json:"info"`     // User's profile.
	Status        int    `db:"status" json:"status"` // User's status
	Password      string `db:"password" json:"-"`
	PasswordTmp   string `db:"passwordtmp" json:"-"`
	Created       int64  `db:"created" json:"created"`
	LastLogin     int64  `db:"-" json:"lastlogin"`
	ActivationKey string `db:"activationkey" json:"-"`
	Type          int    `db:"type" json:"type"` // User's type
}

// IsValidIDFormat 함수는 입력된 아이디가 올바른 형식인지 검사한다.
func IsValidIDFormat(id string) bool {
	if len(id) > IDMaxSize {
		return false
	}

	if !govalidator.IsEmail(id) {
		log.WithFields(log.Fields{
			"func": "IsValidIDFormat",
			"id":   id,
		}).Debug("DEBUG")
		return false
	}

	return true
}

// IsValidPasswordFormat 함수는 입력된 비밀번호가 올바른 형식인지 검사한다.
func IsValidPasswordFormat(password string) bool {
	if len(password) > PasswordMaxSize {
		return false
	}

	return true
}

// Name 함수는 사용자의 이름을 불러온다. 사용자의 이름은 이메일의 @ 앞부분으로 한다.
func (u User) Name() string {
	return strings.Split(u.ID, "@")[0]
}

// IsNormal 함수는 사용자가 정상 가입자인지 여부를 리턴한다.
func (u User) IsNormal() bool {
	return u.Status == UserStatusNormal
}

// IsBlocked 함수는 사용자가 블락 처리 되었는지 여부를 리턴한다.
func (u User) IsBlocked() bool {
	return u.Status == UserStatusBlock
}

// IsWithdraw 함수는 사용자가 탈퇴 상태인지 여부를 리턴한다.
func (u User) IsWithdraw() bool {
	return u.Status == UserStatusWithdraw
}

// IsDeactivated 함수는 사용자가 비활성화(아직 이메일 인증 전) 상태인지 여부를 리턴한다.
func (u User) IsDeactivated() bool {
	return u.Status == UserStatusDeactivated
}

// IsAdmin 함수는 사용자가 운영자인지 여부를 리턴한다.
// 운영자 계정을 발급하려면 가입 후 데이터베이스 관리자가 해당 유저의 Type을 UserTypeAdmin(-1)로
// 수정해야 한다.
func (u User) IsAdmin() bool {
	return u.Type == UserTypeAdmin
}

// Block 함수는 해당 사용자를 블럭 처리 시킨다.
// 블럭 처리된 사용자는 로그인을 할 수 없으며, 로그인을 해야만 사용 가능한 API를 호출할 수 없다.
func (u User) Block(reason string) error {
	db := Database()
	u.Status = UserStatusBlock
	if _, err := db.Auth.Update(&u); err != nil {
		log.Panicf("block failed. id=%v, reason=%v, err=%v", u.ID, reason, err)
		return err
	}
	log.WithFields(log.Fields{
		"id":     u.ID,
		"reason": reason,
	}).Info("BLOCK")

	return nil
}

// UnBlock 함수는 블럭된 해당 사용자를 정상 유저로 변경한다.
// 만일 해당 유저가 블럭된 계쩡이 아니라면 아무 처리도 하지 않는다.
func (u User) UnBlock() error {
	db := Database()
	if u.Status != UserStatusBlock {
		return nil
	}
	u.Status = UserStatusNormal
	if _, err := db.Auth.Update(&u); err != nil {
		log.Panicf("unblock failed. id=%v, err=%v", u.ID, err)
	}
	log.WithFields(log.Fields{"id": u.ID}).Info("UNBLOCK")
	return nil
}

// LoadUserFromID 함수는 사용자의 id를 사용해 데이터베이스로부터 사용자를 읽는다.
func LoadUserFromID(id string) (*User, error) {
	db := Database()
	var user User
	err := db.Auth.SelectOne(&user, "select * from users where id=?", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// LoadUserFromUID 함수는 사용자의 uid를 사용해 데이터베이스로부터 사용자를 읽는다.
func LoadUserFromUID(uid int64) (*User, error) {
	db := Database()
	var user User
	err := db.Auth.SelectOne(&user, "select * from users where uid=?", uid)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// LoadUserFromActivateionKey 함수는 사용자의 ActivationKey를 사용해 데이터베이스로부터
// 사용자를 읽는다.
func LoadUserFromActivateionKey(key string) (*User, error) {
	db := Database()
	var user User
	err := db.Auth.SelectOne(&user, "select * from users where activationkey=?", key)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SendMail 함수는 유저에게 titleTemplate 파일을 제목으로 하고 textTemplate를 내용으로 하는 메일을 보낸다.
func SendMail(user *User, titleTemplate string, textTemplate string) error {
	conf := Config()

	var sbody bytes.Buffer
	stmpl := conf.TemplatePath(titleTemplate)
	utils.JoinTemplate(&sbody, stmpl, user)
	// 이메일 제목에 \n이 있으면 안된다.
	subject := strings.Replace(sbody.String(), "\n", "", -1)

	smtp := &utils.SMTP{
		Server:   conf.SMTP.Server,
		Port:     conf.SMTP.Port,
		User:     conf.SMTP.User,
		Password: conf.SMTP.Password,
	}

	mail := &utils.Email{
		From: mail.Address{
			Name:    conf.SMTP.UserName,
			Address: conf.SMTP.User,
		},
		To: mail.Address{
			Name:    user.Name(),
			Address: user.ID,
		},
		Subject: subject,
	}

	var body bytes.Buffer
	tmpl := conf.TemplatePath(textTemplate)
	utils.JoinTemplate(&body, tmpl, user)
	mail.Body = body.String()

	if err := smtp.Send(mail); err != nil {
		return err
	}

	return nil
}

// 사용자 스키마 상수 정의. 이곳에서 사용되는 상수는 데이터베이스에 반영되므로 값을 변경하면 안된다.
// 불가피하게 값을 변경해야 할 경우는 기존 데이터베이스가 마이그레이션 되어야 한다.
const (
	UserStatusWithdraw    = -1 // 탈퇴 상태
	UserStatusBlock       = -2 // 블럭 상태
	UserStatusDeactivated = 0  // 비활성화(이메일 인증 전) 상태
	UserStatusNormal      = 1  // 정상 상태

	UserTypeAdmin  = -1 // 관리자
	UserTypeNormal = 0  // 일반 사용자

	SaltMaxSize          = 16  // 사용자 비밀번호 필드 중에서 Salt가 차지하는 길이
	IDMaxSize            = 200 // 사용자 아이디 최대 길이
	InfoMaxSize          = 500 // 사용자 정보 최대 길이
	PasswordMaxSize      = 32  // 비밀번호 최대 길이
	ActivationKeyMaxSize = 36  // 이메일 인증키(UUID) 길이
)

func createUserTable(dbmap *gorp.DbMap) {
	gob.Register(&User{})
	table := dbmap.AddTableWithName(User{}, "users").SetKeys(true, "UID")
	table.ColMap("ID").SetMaxSize(IDMaxSize)
	table.ColMap("Info").SetMaxSize(InfoMaxSize)
	table.ColMap("Password").SetMaxSize(SaltMaxSize + PasswordMaxSize)
	table.ColMap("PasswordTmp").SetMaxSize(SaltMaxSize + PasswordMaxSize)
	table.ColMap("ActivationKey").SetMaxSize(ActivationKeyMaxSize)
	table.ColMap("ID").SetUnique(true)
}
