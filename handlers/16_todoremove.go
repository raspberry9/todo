package handlers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"jsproj.com/koo/server/auth/schema"
)

type qTodoRemove struct {
	TID int64 `json:"tid"`
}

type rTodoRemove struct {
	Res int    `json:"res"`
	Msg string `json:"msg"`
}

const (
	todoRemoveOK            = 0
	todoRemoveBadRequest    = -1610
	todoRemoveServerError   = -1620
	todoRemoveNoPermission  = -1630
	todoRemoveDatabaseError = -1640
)

var todoRemoveErrors = map[int]string{
	defaultError: "Error occured during remove a todo.",

	todoRemoveNoPermission: "You might not have permission to remove this todo.",
}

func todoRemoveError(res int) rTodoRemove {
	msg, ok := todoRemoveErrors[res]
	if !ok {
		msg = todoRemoveErrors[defaultError]
	}
	return rTodoRemove{res, msg}
}

// todoRemoveHandler 함수는 사용자가 요청한 일정을 삭제합니다.
func todoRemoveHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {
	var req qTodoRemove
	Unmarshal(r, &req)

	old, err := schema.LoadTodoFromTID(req.TID)
	if err != nil {
		log.Debug(err)
		return todoRemoveError(todoRemoveDatabaseError)
	}

	// 다른 유저의 일정을 조작할 수 없도록 자신의 일정인지 먼저 확인한다.
	if old.OwnerUID != env.Me.UID {
		log.Debugf("old.OwnerUID(%v) is not matched env.Me.UID(%v). maybe hacked.", old.OwnerUID, env.Me.UID)
		return todoRemoveError(todoRemoveNoPermission)
	}

	if err := schema.RemoveTodoFromTID(req.TID); err != nil {
		log.Debug(err)
		return todoRemoveError(todoRemoveDatabaseError)
	}

	return rTodoRemove{todoRemoveOK, "success"}
}
