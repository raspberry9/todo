package schema

import (
	"encoding/gob"

	"gopkg.in/gorp.v1"
)

// Todo 객체는 사용자의 할일 스키마 객체이다. 여기에서 정의된 형태로 데이터베이스 테이블이 작성된다.
// 따라서 이 코드를 바꾸는 경우 반드시 DB 마이그레이션이 필요하다.
type Todo struct {
	TID       int64  `db:"tid" json:"tid"`             // Todo id
	OwnerUID  int64  `db:"owneruid" json:"owneruid"`   // 일정 소유자
	Category  string `db:"category" json:"category"`   // 개인, 쇼핑, 회의, ...
	Todo      string `db:"todo" json:"todo"`           // 생략 불가, 할일
	LimitTime int64  `db:"limittime" json:"limittime"` // 0이면 무기한
	Status    int32  `db:"status" json:"status"`       // 상태 0:진행중, 1:완료됨
}

// Todo 스키마 상수 정의. 이곳에서 사용되는 상수는 데이터베이스에 반영되므로 값을 변경하면 안된다.
// 불가피하게 값을 변경해야 할 경우는 기존 데이터베이스가 마이그레이션 되어야 한다.
const (
	CategoryMaxSize = 10
	TodoMaxSize     = 200

	TodoStatusNormal = 0
	TodoStatusDone   = 1
)

// LoadTodoFromTID 함수는 tid(TODO ID)를 사용해 데이터베이스로부터 Todo 1개를 읽는다.
func LoadTodoFromTID(tid int64) (*Todo, error) {
	db := Database()
	var todo Todo
	err := db.Auth.SelectOne(&todo, "select * from todo where tid=?", tid)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

// LoadTodoListFromTime 함수는 데이터베이스로부터 해당 기간동안의 Todo를 리스트로 읽는다.
func LoadTodoListFromTime(uid int64, stime int64, etime int64) ([]*Todo, error) {
	db := Database()

	var tl []*Todo
	_, err := db.Auth.Select(&tl, "select * from todo where owneruid=? and limittime between ? and ?", uid, stime, etime)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

// RemoveTodoFromTID 함수는 tid(TODO ID)를 사용해 데이터베이스로부터 Todo 1개를 삭제한다.
func RemoveTodoFromTID(tid int64) error {
	db := Database()
	_, err := db.Auth.Exec("delete from todo where tid=?", tid)
	return err
}

func createTodoListTable(dbmap *gorp.DbMap) {
	gob.Register(&Todo{})
	table := dbmap.AddTableWithName(Todo{}, "todo").SetKeys(true, "TID")
	table.ColMap("Category").SetMaxSize(CategoryMaxSize)
	table.ColMap("Todo").SetMaxSize(TodoMaxSize)
}
