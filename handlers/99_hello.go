package handlers

import (
	"net/http"
)

type qHello struct {
}

type rHello struct {
	Country string `json:"country"`
	Price   int    `json:"price"`
}

// helloHandler 함수는 angularjs와 연동 테스트를 위한 함수다.
func helloHandler(w http.ResponseWriter, r *http.Request, env *Environ) interface{} {
	var req qHello
	Unmarshal(r, &req)

	// 테스트를 위한 객체 리스트를 생성한다.
	hellos := []rHello{}
	hellos = append(hellos, rHello{"Korea", 101})
	hellos = append(hellos, rHello{"Japan", 202})
	hellos = append(hellos, rHello{"China", 303})
	hellos = append(hellos, rHello{"USA", 404})

	return hellos
}
