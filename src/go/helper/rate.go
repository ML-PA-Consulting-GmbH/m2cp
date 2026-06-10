package helper

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"m2cpcli/graphql"
)

func FindRatingFromInputRating(ratings *[]graphql.SnapRating, rating string) (*graphql.SnapRating, error) {
	capitalizedRating := cases.Title(language.English, cases.NoLower).String(rating)
	for _, r := range *ratings {
		if r.Name == capitalizedRating {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("rating %s not found", rating)
}
