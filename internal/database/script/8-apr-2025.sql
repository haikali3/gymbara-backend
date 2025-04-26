SELECT * FROM users;
SELECT * FROM subscriptions;
SELECT * FROM exercises;
SELECT * FROM Exercisedetails;
SELECT * FROM workoutsections;

DELETE FROM Subscriptions WHERE user_id = 1;
UPDATE Users SET stripe_customer_id = NULL, is_premium = FALSE WHERE id = 1;

SELECT stripe_subscription_id, expiration_date 
	 FROM Subscriptions 
	 WHERE user_id = (SELECT id FROM Users WHERE email = 'manfdvcl9@gmail.com') 
	 AND expiration_date > NOW()
	 ORDER BY expiration_date DESC
	 LIMIT 1

SELECT * FROM userExercisesDetails
ORDER BY submitted_at DESC
