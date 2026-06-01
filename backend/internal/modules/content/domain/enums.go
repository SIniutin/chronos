package domain

type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusUpdating  ContentStatus = "updating"
	ContentStatusArchived  ContentStatus = "archived"
)

func (s ContentStatus) IsPublished() bool {
	return s == ContentStatusPublished
}

type ChallengeType string

const (
	ChallengeTypeTheory       ChallengeType = "theory"
	ChallengeTypeSingleChoice ChallengeType = "single_choice"
	ChallengeTypeMultiple     ChallengeType = "multiple_choice"
	ChallengeTypeTimeline     ChallengeType = "timeline"
	ChallengeTypeMatchPairs   ChallengeType = "match_pairs"
	ChallengeTypeImage        ChallengeType = "image_question"
	ChallengeTypeMatchImage   ChallengeType = "match_image"
	ChallengeTypeMatchPhotos  ChallengeType = "match_photos"
	ChallengeTypeQuote        ChallengeType = "quote_question"
	ChallengeTypeTrueFalse    ChallengeType = "true_false"
	ChallengeTypeFillBlank    ChallengeType = "fill_in_blank"
	ChallengeTypeMapPoint     ChallengeType = "map_point"
	ChallengeTypeMapArea      ChallengeType = "map_area"
)

func NewChallengeType(raw string) (ChallengeType, error) {
	t := ChallengeType(raw)
	switch t {
	case ChallengeTypeTheory, ChallengeTypeSingleChoice, ChallengeTypeMultiple, ChallengeTypeTimeline,
		ChallengeTypeMatchPairs, ChallengeTypeImage, ChallengeTypeMatchImage, ChallengeTypeQuote,
		ChallengeTypeMatchPhotos, ChallengeTypeTrueFalse, ChallengeTypeFillBlank, ChallengeTypeMapPoint, ChallengeTypeMapArea:
		return t, nil
	default:
		return "", ErrInvalidInput
	}
}

type Difficulty string

const (
	DifficultyUndefined Difficulty = "undefined"
	DifficultyEasy      Difficulty = "easy"
	DifficultyMedium    Difficulty = "medium"
	DifficultyHard      Difficulty = "hard"
)

func NewDifficulty(raw string) (Difficulty, error) {
	d := Difficulty(raw)
	switch d {
	case DifficultyEasy, DifficultyMedium, DifficultyHard, DifficultyUndefined:
		return d, nil
	default:
		return "", ErrInvalidInput
	}
}
