package schema

import (
	"database/sql"

	log "github.com/Sirupsen/logrus"
	"github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v1"
)

// Databases 구조체는 데이터베이스의 묶음 객체이다.
type Databases struct {
	Auth *gorp.DbMap
}

var (
	dbs Databases
)

// IsDuplicated 함수는 데이터베이스 Insert 쿼리 결과 error 를 확인하여 중복오류가 발생 했는지
// 여부를 리턴한다.
func (d *Databases) IsDuplicated(err error) bool {
	if err != nil {
		if err.(*mysql.MySQLError).Number == 1062 {
			return true
		}
	}
	return false
}

func mustInitDatabase(conf *Configure) {
	auth := mustInitAuthDatabase(conf.Database.Auth)
	dbs = Databases{Auth: auth}
	log.Info("databases initialized.")
}

// Database 함수는 현재 생성된 Databases 구조체의 인스턴스를 반환한다.
func Database() *Databases {
	return &dbs
}

func mustInitAuthDatabase(dsn string) *gorp.DbMap {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("db open error. dsn=%s, err=%v", dsn, err)
	}

	dbmap := gorp.DbMap{
		Db: db, Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8"}}

	// 테이블 생성
	createUserTable(&dbmap)
	createTodoListTable(&dbmap)

	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		log.Fatalf("table create error. dsn=%s, err=%v", dsn, err)
	}

	return &dbmap
}
