package dataloader

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func fetchUserIDs(db *gorm.DB, tokens []string) ([]int, error) {
	rows, err := db.Table("access_tokens").Select("distinct user_id").Where("expire > ?", time.Now()).Where("value IN (?)", tokens).Rows()
	if err != nil {
		return nil, err
	}
	userIDs := make([]int, 0)
	for rows.Next() {
		var id int
		if err := db.ScanRows(rows, &id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	rows.Close()

	return userIDs, nil
}

func getUserMap(db *gorm.DB, userIDs []int) (map[int]*models.User, error) {
	var userMap map[int]*models.User
	if len(userIDs) > 0 {

		var users []*models.User
		if err := db.Where("id IN (?)", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}

		userMap = make(map[int]*models.User, len(users))
		for _, user := range users {
			userMap[user.ID] = user
		}
	} else {
		userMap = make(map[int]*models.User, 0)
	}

	return userMap, nil
}

func getTokenMap(db *gorm.DB, tokens []string) (map[string]*models.AccessToken, error) {
	var accessTokens []*models.AccessToken
	err := db.Where("expire > ?", time.Now()).Where("value IN (?)", tokens).Find(&accessTokens).Error
	if err != nil {
		return nil, err
	}

	tokenMap := make(map[string]*models.AccessToken, len(tokens))

	for _, token := range accessTokens {
		tokenMap[token.Value] = token
	}
	return tokenMap, nil
}

func getUsers(tokens []string, tokenMap map[string]*models.AccessToken, userMap map[int]*models.User) []*models.User {
	result := make([]*models.User, len(tokens))
	for i, token := range tokens {
		accessToken, tokenFound := tokenMap[token]
		if tokenFound {
			user, userFound := userMap[accessToken.UserID]
			if userFound {
				result[i] = user
			}
		}
	}

	return result
}

func NewUserLoaderByToken(db *gorm.DB) *UserLoader {
	return &UserLoader{
		maxBatch: 100,
		wait:     5 * time.Millisecond,
		fetch: func(tokens []string) ([]*models.User, []error) {
			userIDs, err := fetchUserIDs(db, tokens)
			if err != nil {
				return nil, []error{err}
			}

			userMap, err := getUserMap(db, userIDs)
			if err != nil {
				return nil, []error{err}
			}

			tokenMap, err := getTokenMap(db, tokens)
			if err != nil {
				return nil, []error{err}
			}

			result := getUsers(tokens, tokenMap, userMap)

			return result, nil
		},
	}
}
