package routes

import (
	"fmt"
	"membuatuser/controller"
	"membuatuser/db"
	middleware "membuatuser/middleware"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

)

func Init() error {
	e := echo.New()
	db, err := db.Init()
	if err != nil {
		return err
	}
	defer db.Close()
	//menunda penutupan database => close

	e.GET("", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Application is Running",
		})
	})

	task := e.Group("/task")

	task.Use(middleware.ValidateToken)

	task.GET("", controller.GetTasksController(db)) //=>untuk mengirimkan db
	task.GET("/:id", controller.GetTaskByIDController(db))
	task.POST("/add", controller.AddTaskController(db))
	task.POST("", controller.SearchTasksFormController(db))
	e.POST("/register", controller.RegisterController(db))
	e.POST("/login", controller.LoginCompareController(db))
	e.POST("/logout", controller.LogoutController(db))
	task.PUT("/:id", controller.EditTaskController(db))
	task.DELETE("/:id", controller.DeleteTaskController(db))
	task.DELETE("", controller.BulkDeleteController(db))
	task.GET("/status", controller.CountTexts(db))
	return e.Start(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")))
}
