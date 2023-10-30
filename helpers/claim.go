package helpers

import (
	"membuattodo/middleware"

	"github.com/labstack/echo/v4"
)

func ClaimToken(c echo.Context) (response middleware.JWTClaim) {
	user := c.Get("jwt-res")
	response = user.(middleware.JWTClaim)
	return
}
