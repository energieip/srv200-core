package api

import (
	"encoding/json"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/context"
	"github.com/mitchellh/mapstructure"
)

func (api *API) createToken(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	var creds Credentials
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	user := database.GetUser(api.db, creds.UserKey)
	if user == nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized access", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(TokenExpirationTime * time.Second)
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(api.apiPassword))
	if err != nil {
		api.sendError(w, APIErrorInvalidValue, "Error during token generation", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     TokenName,
		Value:    tokenString,
		Expires:  expirationTime,
		MaxAge:   TokenExpirationTime,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
		Path:     "/",
	})

	res := JwtToken{
		Token:     tokenString,
		TokenType: "bearer",
		ExpireIn:  TokenExpirationTime,
	}
	api.access.Set(tokenString, *user)
	json.NewEncoder(w).Encode(res)
}

func (api *API) getUserInfo(w http.ResponseWriter, req *http.Request) {
	decoded := context.Get(req, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)

	json.NewEncoder(w).Encode(auth)
}

func (api *API) logout(w http.ResponseWriter, req *http.Request) {
	decoded := context.Get(req, "token")
	var tokenString string
	mapstructure.Decode(decoded.(string), &tokenString)

	// see https://golang.org/pkg/net/http/#Cookie
	// Setting MaxAge<0 means delete cookie now.

	http.SetCookie(w, &http.Cookie{
		Name:     TokenName,
		MaxAge:   -1,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
		Path:     "/",
	})

	api.access.Remove(tokenString)

	w.Write([]byte("{}"))
}
