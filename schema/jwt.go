package schema

import (
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"jsproj.com/koo/gosari/utils"
)

const (
	tokenExpireHour = 24
)

var (
	privateKey []byte
	publicKey  []byte
)

func mustInitJWT(conf *Configure) {
	var err error
	privateKey, err = ioutil.ReadFile(conf.PrivateKeyFile())
	if err != nil {
		log.Fatalf("invalid private key. err=%v", err)
	}

	publicKey, err = ioutil.ReadFile(conf.PublicKeyFile())
	if err != nil {
		log.Fatalf("invalid public key. err=%v", err)
	}
}

// IssueToken is issue a jason web token(jwt).
// It need rsa key fo generate the token.
//  openssl genrsa -out mykey.rsa 1024
//  openssl rsa -in mykey.rsa -pubout > mykey.rsa.pub
func IssueToken(user *User) (string, error) {
	// Create the token
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	// Set some claims
	token.Claims["uid"] = user.UID
	token.Claims["exp"] = utils.ServerTime() + (3600 * tokenExpireHour)
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseToken implements the jwt token parser.
// It returns user unique id(uid), expire time(exp) and error.
func ParseToken(r *http.Request) (int64, float64, error) {
	token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return 0, 0, err
	}
	if !token.Valid {
		return 0, 0, errors.New("invalid token")
	}
	return int64(token.Claims["uid"].(float64)), token.Claims["exp"].(float64), nil
}

// LoadUserFromRequest is get a user from database using uid of jwt.
func LoadUserFromRequest(r *http.Request) (*User, error) {
	uid, _, err := ParseToken(r)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	user, err := LoadUserFromUID(uid)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	return user, nil
}
