package dataloader

import (
	"math/rand"
	"testing"

	"github.com/joho/godotenv"
	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Test_NewUserLoader(t *testing.T) {
	err := godotenv.Load("../testing.env")

	db, err := database.SetupDatabase()
	if err != nil {
		t.Fatal("failed to connect database")
	}

	if err := db.AutoMigrate(&models.User{}, &models.AccessToken{}); err != nil {
		t.Fatal(err)
	}

	user, err := models.RegisterUser(db, "test"+randomString(10), nil, false)

	if err != nil {
		t.Fatal(err)
	}

	token, err := user.GenerateAccessToken(db)
	if err != nil {
		t.Fatal(err)
	}

	userLoader := NewUserLoaderByToken(db)

	userLoaded, errors := userLoader.fetch([]string{token.Value})

	if errors != nil {
		t.Fatal(err)
	}

	if userLoaded[0].ID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, userLoaded[0].ID)
	}
}
