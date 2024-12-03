package middleware

import (
	"fmt"
	"goTogether/utility"
	"net/http"
	"os"

	"google.golang.org/api/idtoken"
)

type serverHandler func(http.ResponseWriter, *http.Request)

func Verify(next serverHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if token == nil {
			fmt.Println("no token received")
			utility.RespondWithError(w, 400, "No token received")
			return
		}
		fmt.Println("gonna validate token...", token)
		tokenString := token.Value
		if err != nil {
			utility.RespondWithJson(w, 400, "Cookie not found")
			return
		}
		var audience string = os.Getenv("GOOGLE_KEY")
		// var audience string = "737460282691-55n287julvnh30i5empcq91marqh3apq.apps.googleusercontent.com"
		payload, err := idtoken.Validate(r.Context(), tokenString, audience)
		if err != nil {
			utility.RespondWithError(w, 400, "Not valid token")
			return
		}
		claims := payload.Claims

		// type validToken struct {
		// 	Message string `json:"message"`
		// }
		fmt.Println(claims["email"])
		next(w, r)
		// utility.RespondWithJson(w, 200, validToken{Message: "ok"})
	}

}
