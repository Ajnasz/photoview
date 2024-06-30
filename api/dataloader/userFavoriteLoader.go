package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func makeUniqueIDs(idMap map[int]struct{}) []int {
	uniqueIDs := make([]int, len(idMap))
	count := 0
	for id := range idMap {
		uniqueIDs[count] = id
		count++
	}

	return uniqueIDs
}

func isFavorite(userMediaFavorites []*models.UserMediaData, userID int, mediaID int) bool {
	for _, fav := range userMediaFavorites {
		if fav.UserID == userID && fav.MediaID == mediaID {
			return true
		}
	}

	return false
}

func NewUserFavoriteLoader(db *gorm.DB) *UserFavoritesLoader {
	return &UserFavoritesLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: func(keys []*models.UserMediaData) ([]bool, []error) {
			userIDMap := make(map[int]struct{}, len(keys))
			mediaIDMap := make(map[int]struct{}, len(keys))
			for _, key := range keys {
				userIDMap[key.UserID] = struct{}{}
				mediaIDMap[key.MediaID] = struct{}{}
			}

			uniqueUserIDs := makeUniqueIDs(userIDMap)
			uniqueMediaIDs := makeUniqueIDs(mediaIDMap)

			var userMediaFavorites []*models.UserMediaData
			err := db.Where("user_id IN (?)", uniqueUserIDs).Where("media_id IN (?)", uniqueMediaIDs).Where("favorite = TRUE").Find(&userMediaFavorites).Error
			if err != nil {
				return nil, []error{err}
			}

			result := make([]bool, len(keys))
			for i, key := range keys {
				result[i] = isFavorite(userMediaFavorites, key.UserID, key.MediaID)
			}

			return result, nil
		},
	}
}
