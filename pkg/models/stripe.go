package models

type StripeProvider struct {
	SecretKey string
}

type Product struct {
	ProductID string
	PriceID   string
	Price     int64
}
