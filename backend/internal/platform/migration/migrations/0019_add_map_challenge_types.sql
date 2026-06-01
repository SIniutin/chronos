-- +goose Up
ALTER TABLE challenges DROP CONSTRAINT challenges_type_check;
ALTER TABLE challenges
    ADD CONSTRAINT challenges_type_check
    CHECK (challenge_type IN (
        'theory',
        'single_choice',
        'multiple_choice',
        'timeline',
        'match_pairs',
        'image_question',
        'match_image',
        'match_photos',
        'quote_question',
        'true_false',
        'fill_in_blank',
        'map_point',
        'map_area'
    ));

-- +goose Down
ALTER TABLE challenges DROP CONSTRAINT challenges_type_check;
ALTER TABLE challenges
    ADD CONSTRAINT challenges_type_check
    CHECK (challenge_type IN (
        'theory',
        'single_choice',
        'multiple_choice',
        'timeline',
        'match_pairs',
        'image_question',
        'match_image',
        'quote_question',
        'true_false',
        'fill_in_blank'
    ));
