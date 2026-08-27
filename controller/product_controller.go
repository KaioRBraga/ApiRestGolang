package controller

import (
	"net/http"
	"restapi/model"

	"github.com/gin-gonic/gin"
)

type productController struct {
	//usecase usecase.ProductUsecase
}

func NewProductController() productController {
	return productController{}
}

func (p *productController) GetProducts(ctx *gin.Context) {
	product := []model.Product{
		{ID: 1, Name: "Banana", Price: 3.99},
		{ID: 2, Name: "Apple", Price: 2.99},
		{ID: 3, Name: "Orange", Price: 4.99},
	}

	ctx.JSON(http.StatusOK, product)

}
