package middleware

import (
	"membuatuser/db"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

)

type JWTClaim struct {
	ID    int
	Email string
	jwt.StandardClaims
}

func ValidateToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var reqToken string

		headerDataToken := c.Request().Header.Get("Authorization")

		splitToken := strings.Split(headerDataToken, "Bearer ")
		if len(splitToken) > 1 {
			reqToken = splitToken[1]
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		if !CekTokenDatabase(reqToken) {
			return echo.NewHTTPError(http.StatusUnauthorized, 
				"token is not entered",
			)
		}
		
		jwtResponse, err := ClaimsJWT(reqToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, 
				"token is not entered",
			)
		}

		c.Set("jwt-res", jwtResponse)

		return next(c)
	}
}

func ClaimsJWT(token string) (response JWTClaim, err error) {
	var (
		secretKey = []byte("secret")
		jwtClaims JWTClaim
		jwtToken  *jwt.Token
	)

	jwtToken, err = jwt.ParseWithClaims(
		token,
		&jwtClaims,
		func(jwtToken *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)
	if err != nil {
		return
	}

	if jwtToken == nil || !jwtToken.Valid {
		return
	}

	response = JWTClaim{
		ID:    jwtClaims.ID,
		Email: jwtClaims.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwtClaims.ExpiresAt,
		},
	}

	return
}

func CekTokenDatabase(token string) bool {
	var exist bool
	db, err := db.Init()
	if err != nil {
		return false
	}
	defer db.Close()

	query := `SELECT EXISTS (SELECT 1 FROM user_token WHERE token = $1)`
	err = db.Get(&exist, query, token)
	if err != nil {
		return false
	}
	return exist
}
