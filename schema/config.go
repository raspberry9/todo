package schema

import (
	"path/filepath"
	"strings"

	"code.google.com/p/gcfg"
	log "github.com/Sirupsen/logrus"
)

// Configure 구조체는 설정 파일 스키마이다.
type Configure struct {
	Server struct {
		Bind      string `json:"bind"`
		Test      string `json:"test"`
		Whitelist string `json:"whitelist"`
		AppID     string `json:"appid"`
	} `json:"server"`
	Redis struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"redis"`
	Database struct {
		Auth string `json:"auth"`
	} `json:"database"`
	SMTP struct {
		Server   string `json:"server"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		UserName string `json:"username"`
		Password string `json:"password"`
	} `json:"smtp"`
	Activation struct {
		Use string `json:"use"`
	} `json:"activation"`
	Resources struct {
		PrivateKeyFile string `json:"privatekeyfile"`
		PublicKeyFile  string `json:"publickeyfile"`
		TemplatePath   string `json:"templatepath"`
		StaticPath   string `json:"staticpath"`
	} `json:"resources"`
}

var (
	conf     Configure
	conffile string
)

func mustInitConfig(confFileName string) {
	conffile = confFileName
	LoadConfig()
	log.Info("config file loaded.")
}

// LoadConfig 함수는 설정을 읽는다.
func LoadConfig() {
	err := gcfg.ReadFileInto(&conf, conffile)
	if err != nil {
		log.Fatalf("read config error. file=%s, err=%v", conffile, err)
	}
}

// Config 함수는 현재 설정 구조체를 리턴한다.
func Config() *Configure {
	return &conf
}

// IsAllowIP 함수는 해당 IP가 화이트리스트에 포함 되었는지 여부를 리턴한다.
// 화이트 리스트에 IP를 추가 하려면 [server] 섹션의 whitelist 항목에 추가한다.
// 주의할 점은 인자로 넣는 IP가 nginx등으로 포워딩 되어 들어오는 경우 서버의 IP로 바뀔 수 있기 때문에
// http.Request.RemoteAddr을 사용하지 말고 handlers.GetIP(r) 함수를 사용해야 한다.
func (c *Configure) IsAllowIP(remoteAddr string) bool {
	for _, element := range strings.Split(c.Server.Whitelist, ",") {
		if strings.Trim(element, " ") == strings.Split(remoteAddr, ":")[0] {
			return true
		}
	}
	return false
}

// TemplatePath 함수는 인자로 주어진 파일명으로부터 템플릿 파일의 전체 경로를 얻어온다.
// 템플릿 파일의 경로는 [resources] 섹션의 templatepath 항목에서 설정한다.
func (c *Configure) TemplatePath(filename string) string {
	p, err := filepath.Rel("", c.Resources.TemplatePath)
	if err != nil {
		log.Errorf("template path error. path=%s, file=%s", p, filename)
	}
	return p + "/" + filename
}

// PrivateKeyFile 함수는 jwt private key의 경로를 반환한다.
func (c *Configure) PrivateKeyFile() string {
	return c.Resources.PrivateKeyFile
}

// PublicKeyFile 함수는 jwt public key의 경로를 반환한다.
func (c *Configure) PublicKeyFile() string {
	return c.Resources.PublicKeyFile
}

// StaticPath 함수는 인자로 주어진 파일명으로부터 템플릿 파일의 전체 경로를 얻어온다.
// 템플릿 파일의 경로는 [resources] 섹션의 staticpath 항목에서 설정한다.
func (c *Configure) StaticPath(filename string) string {
	p, err := filepath.Rel("", c.Resources.StaticPath)
	if err != nil {
		log.Errorf("static path error. path=%s, file=%s", p, filename)
	}
	return p + "/" + filename
}

// IsTestServer 함수는 테스트 서버 여부를 설정 파일로부터 반환한다.
func (c *Configure) IsTestServer() bool {
	return strings.EqualFold(c.Server.Test, "true")
}

// IsUseActivation 함수는 이메일 인증 사용 여부를 설정 파일로부터 반환한다.
func (c *Configure) IsUseActivation() bool {
	return strings.EqualFold(c.Activation.Use, "true")
}
