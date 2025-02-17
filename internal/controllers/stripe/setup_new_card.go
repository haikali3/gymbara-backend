// https://snassr.hashnode.dev/go-and-stripe-subscriptions-quickstart

package controllers

import (
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/setupintent"
)

func SetupNewCard(customerID string) (secret *string, err error) {
	params := &stripe.SetupIntentParams{
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
		Customer: &customerID,
	}
	si, err := setupintent.New(params)
	if err != nil {
		return nil, err
	}

	return &si.ClientSecret, nil
}
