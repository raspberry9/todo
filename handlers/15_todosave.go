package handlers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"jsproj.com/koo/server/auth/schema"
)

type qTodoSave struct {
	TID       int64  `json:"tid"`
	LimitTime int64  `json:"limittime"`
	Category  string `json:"category"`
	Todo      string `json:"todo"`
}

type rTodoSave struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

const (
	todoSaveOK            = 0
	todoSaveBadRequest    = -1510
	todoSaveServerError   = -1520
	todoSaveNoPermission  = -1530
	todoSaveDatabaseError = -1540
)

var todoSaveErrors = map[int]string{
	defaultError: "Error occured during todo list.",

	todoSaveNoPermission: "You might not have permission to save this todo.",
}

func todoSaveError(res int) rTodoSave {
	msg, ok := todoSaveErrors[res]
	if !ok {
		msg = todoSaveErrors[defaultError]
	}
	return rTodoSave{res, msg}
}

// todoSaveHandler 함수는 사용자가 요청한 일정을 생성 혹은 업데이트 합니다.
func todoSaveHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {
	var req qTodoSave
	Unmarshal(r, &req)

	todo := schema.Todo{
		TID:       req.TID,
		OwnerUID:  env.Me.UID,
		Category:  req.Category,
		Todo:      req.Todo,
		LimitTime: req.LimitTime,
		Status:    schema.TodoStatusNormal}

	if isNew := todo.TID == 0; isNew == true {
		if err := env.DB.Auth.Insert(&todo); err != nil {
			log.Debug(err)
			return todoSaveError(todoSaveDatabaseError)
		}
	} else {
		old, err := schema.LoadTodoFromTID(todo.TID)
		if err != nil {
			log.Debug(err)
			return todoSaveError(todoSaveDatabaseError)
		}

		// 다른 유저의 일정을 조작할 수 없도록 자신의 일정인지 먼저 확인한다.
		if old.OwnerUID != todo.OwnerUID {
			log.Debugf("old.OwnerUID(%v) is not matched todo.OwnerUID(%v). maybe hacked.", old.OwnerUID, todo.OwnerUID)
			return todoSaveError(todoSaveNoPermission)
		}

		if _, err := env.DB.Auth.Update(&todo); err != nil {
			log.Debug(err)
			return todoSaveError(todoSaveDatabaseError)
		}
	}

	return rTodoSave{todoSaveOK, "success"}
}
