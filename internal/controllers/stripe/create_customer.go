// https://snassr.hashnode.dev/go-and-stripe-subscriptions-quickstart

package controllers

import (
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
)

func CreateCustomer(name, email, desc string) (*string, error) {
	params := &stripe.CustomerParams{
		Name:        &name,
		Email:       &email,
		Description: &desc,
	}
	cust, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return &cust.ID, nil
}
