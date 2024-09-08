package cookie

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

const jwtKey = "R@ndomSecretKey"

func NewToken(value string, expirationTime time.Time) (string, *Claims, error) {
	claims := &Claims{
		// plain text string
		ID: value,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", nil, err
	}
	return tokenString, claims, nil
}

func CheckToken(tokenStr string) (*Claims, bool, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		return &Claims{}, false, err
	}
	// expired
	if !token.Valid {
		return &Claims{}, false, nil
	}
	return claims, true, nil
}

func RefreshToken(expirationTime time.Time, tokenStr string) (string, bool, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return "", false, err
	}
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", false, err
	}
	return tokenString, true, nil
}

func NewCookie(w http.ResponseWriter, value string) (map[string]string, error) {
	res := make(map[string]string)
	expirationTime := time.Now().Add(5 * time.Minute)
	tokenString, claims, err := NewToken(value, expirationTime)
	if err != nil {
		return nil, err
	}
	// encrypted token
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	// plain text id
	http.SetCookie(w, &http.Cookie{
		Name: "id",
		// claims.ID contains `value`
		Value:   claims.ID,
		Expires: expirationTime,
	})
	res["id"] = (*claims).ID
	res["token"] = tokenString
	return res, nil
}

func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func ValidateCookie(r *http.Request) error {
	// plain text
	ID, err := GetCookie(r, "id")
	if err != nil {
		return err
	}
	// encoded
	tokenString, err := GetCookie(r, "token")
	if err != nil {
		return err
	}
	// decode token string
	claims, ok, err := CheckToken(tokenString)
	if err != nil {
		return err
	}
	if !ok {
		err := errors.New("token expired")
		return err
	}
	// compare
	if claims.ID != ID {
		err := errors.New("token invalid")
		return err
	}
	return nil
}

func EnsureCookie(w http.ResponseWriter, r *http.Request, value string) (string, error) {
	//`id` -- hardcoded key
	// contains plain text value
	idValue, err := GetCookie(r, "id")
	if err != nil {
		// No cookie found
		// Create new cookie
		// cookie contains:
		// -- id: <plain_text_value>
		// -- token: <encrypted jwt token>
		myCookies, err := NewCookie(w, value)
		if err != nil {
			return "", err
		}
		// return plain text part
		idValue = myCookies["id"]
	}
	// return cookie value
	return idValue, nil
}
