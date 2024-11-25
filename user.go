package main

import (
	"time"
	"math/rand"
)


// Id and CreatedAt are never used in the code, maybe can get rid of them?
type User struct {
	Id				int			`field:"id"`
	Pets			int			`field:"pets"`
	UserID			string		`field:"user_id"`
	DisplayName		string		`field:"display_name"`
	CreatedAt		time.Time	`field:"created_at"`
}

func generateRandomName() string {
	adjectives := []string{"happy","sad","mad"}
	nouns := []string{"Cat","Dog","Henry"}

	adj_i := rand.Intn(len(adjectives))
	noun_i := rand.Intn(len(nouns))

	return adjectives[adj_i] + nouns[noun_i]

}


func CreateUser(uid string) User {
	var newUser User

	newUser.DisplayName = generateRandomName()
	newUser.UserID = uid

	return newUser
}

