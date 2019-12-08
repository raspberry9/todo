// Package schema implements data structure of authentication http server.
package schema

// MustInit 함수는 schema 패키지 전체를 초기화 한다.
func MustInit(configFileName string) {
	mustInitConfig(configFileName)
	mustInitDatabase(Config())
	mustInitJWT(Config())
}
