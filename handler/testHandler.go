package handler

import (
	"net/http"

	"goTogether/utility"
)

func Testhandler(w http.ResponseWriter, r *http.Request) {
	type okResponse struct {
		Message string `json:"message"`
	}
	okresponse := okResponse{Message: "server is running"}
	utility.RespondWithJson(w, 200, okresponse)
}
