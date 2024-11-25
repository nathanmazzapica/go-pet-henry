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
	adjectives := []string{"big", "long", "small", "golden", "yellow", "black",
							"red", "short", "cunning", "silly","radical","sluggish",
							"speedy","humorous","shy","scared","brave","intelligent","stupid"}

	nouns := []string{"Dog","Watermelon","Crusader","Lancer","Envisage","Frog",
					"Beetle","Cellphone","Python","Lizard","Butterfly","Dragon",
					"Automobile","Cow","Henry","Levi","Array","Buzzer","Balloon"}

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

