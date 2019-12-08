package handlers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"jsproj.com/koo/server/auth/schema"
)

type qTodoList struct {
	Sdate int64 `json:"sdate"`
	Edate int64 `json:"edate"`
}

type rTodoList struct {
	Res      int            `json:"res"`
	Msg      string         `json:"msg"`
	TodoList []*schema.Todo `json:"todolist"`
}

const (
	todoListOK          = 0
	todoListBadRequest  = -1410
	todoListServerError = -1420
)

var todoListErrors = map[int]string{
	defaultError: "Error occured during todo list.",
}

func todoListError(res int) rTodoList {
	msg, ok := todoListErrors[res]
	if !ok {
		msg = todoListErrors[defaultError]
	}
	return rTodoList{res, msg, nil}
}

// todolistHandler 함수는 사용자의 일정 목록을 반환합니다.
func todolistHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {
	var req qTodoList
	Unmarshal(r, &req)

	tl, err := schema.LoadTodoListFromTime(env.Me.UID, req.Sdate, req.Edate)
	if err != nil {
		log.Debugf("[todolistHandler] tl=%v, err=%v", tl, err)
		return todoListError(todoListServerError)
	}

	return rTodoList{todoListOK, "success", tl}
}
