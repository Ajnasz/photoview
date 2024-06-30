package dataloader

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

func Test_NewUserFavoriteLoader(t *testing.T) {
	err := godotenv.Load("../testing.env")

	db, err := database.SetupDatabase()
	if err != nil {
		t.Fatal("failed to connect database")
	}
	if err := db.AutoMigrate(&models.UserMediaData{}); err != nil {
		t.Fatal(err)
	}

	user, err := models.RegisterUser(db, "test"+randomString(10), nil, false)
	if err != nil {
		t.Fatal(err)
	}

	userMediaData := models.UserMediaData{
		UserID:   user.ID,
		MediaID:  1,
		Favorite: true,
	}

	if result := db.Create(&userMediaData); result.Error != nil {
		t.Fatal(result.Error)
	}

	userFavoriteLoader := NewUserFavoriteLoader(db)

	data, errors := userFavoriteLoader.fetch([]*models.UserMediaData{
		{
			UserID:  1,
			MediaID: 1,
		},
	})

	if errors != nil {
		t.Fatal(errors)
	}

	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}

	if !data[0] {
		t.Fatalf("expected true, got false")
	}
}
