package main

import (
	"restapi/controller"
	"restapi/db"
	"restapi/repository"
	"restapi/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	dbConnection, err := db.ConnectDB()
	if err != nil {
		panic(err)
	}

	//Camada para repository
	repository := repository.NewProductRepository(dbConnection)

	//Camada para usecase
	ProductUsecase := usecase.NewProductUsecase(repository)
	//camada para controller
	ProductController := controller.NewProductController(ProductUsecase)

	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	server.GET("/products", ProductController.GetProducts)

	server.Run(":8000")
}
