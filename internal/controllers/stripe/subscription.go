package controllers

import (
	"fmt"

	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/subscriptionitem"
)

func CreateSubscription(customerID, priceID, paymentMethodID string, trialEnd int64) (*string, error) {
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: &customerID,
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: &priceID,
			},
		},
		TrialEnd:             &trialEnd,
		DefaultPaymentMethod: &paymentMethodID,
	}
	sb, err := subscription.New(subscriptionParams)
	if err != nil {
		return nil, err
	}

	return &sb.ID, nil
}

func UpdateSubscription(subscriptionID, priceID string) (*string, error) {
	subItemParams := &stripe.SubscriptionItemListParams{
		Subscription: &subscriptionID,
	}
	i := subscriptionitem.List(subItemParams)
	var si *stripe.SubscriptionItem
	for i.Next() {
		si = i.SubscriptionItem()
		break
	}

	if si == nil {
		return nil, fmt.Errorf("no subscription items found for subscription %v", subscriptionID)
	}

	subsciptionParams := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
		ProrationBehavior: stripe.String(string(stripe.SubscriptionSchedulePhaseProrationBehaviorCreateProrations)),

		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    &si.ID,
				Price: &priceID,
			},
		},
	}

	stripeSubscription, err := subscription.Update(subscriptionID, subsciptionParams)
	if err != nil {
		return nil, err
	}

	return &stripeSubscription.ID, nil
}

func CancelSubscription(subscriptionID string) error {
	cancel := true
	subscriptionParams := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: &cancel,
	}

	_, err := subscription.Update(subscriptionID, subscriptionParams)
	if err != nil {
		return err
	}

	return nil
}
