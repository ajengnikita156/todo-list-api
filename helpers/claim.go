package helpers

import (
	"github.com/labstack/echo/v4"

	"membuatuser/middleware"
)

func ClaimToken(c echo.Context) (response middleware.JWTClaim) {
	user := c.Get("jwt-res")
	response = user.(middleware.JWTClaim)
	return
}