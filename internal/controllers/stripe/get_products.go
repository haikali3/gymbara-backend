package controllers

import (
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/price"
)

func GetProducts() ([]models.Product, error) {
	products := make([]models.Product, 0)

	priceParams := &stripe.PriceListParams{}
	priceIterator := price.List(priceParams)
	for priceIterator.Next() {
		products = append(products, models.Product{
			ProductID: priceIterator.Price().Product.ID,
			PriceID:   priceIterator.Price().ID,
			Price:     priceIterator.Price().UnitAmount,
		})
	}

	return products, nil
}
