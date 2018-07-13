package mvc

import (
	"github.com/labstack/echo"
	"github.com/lpisces/postcode/cmd/serve/mvc/c"
)

func Route(e *echo.Echo) {

	// home
	e.GET("/", c.GetHome)

	// csv
	e.GET("/csv", c.GetCSV)

}
