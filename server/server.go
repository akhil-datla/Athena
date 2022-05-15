package server

import (
	"fmt"
	"main/backend"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

//e is the Echo Server Object
var e = echo.New()

//StartServer starts the server
func StartServer(port int) {
	e.HideBanner = true
	InitializeRoutes()
	e.Use(middleware.Recover())
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

//InitializeRoutes initializes the routes
func InitializeRoutes() {
	e.GET("/search", func(c echo.Context) error {
		query := c.QueryParam("query")
		results := backend.GetQuestionsAndAnswers(backend.RankedQuerySearch(query))
		return c.JSON(200, results)
	})
}
