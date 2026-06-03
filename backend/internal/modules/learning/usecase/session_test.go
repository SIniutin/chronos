package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	learning_api "github.com/SIniutin/history-app-backend/internal/modules/learning/api"
	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/google/uuid"
)

func TestStartSessionUsesDefaultLimitAndStoresQueue(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	repo := newLearningMemoryRepo()
	content := &contentMemoryRepo{challenges: makeChallenges(skillID, 12)}
	uc := NewServiceFromRepository(repo, content)

	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID:  userID.String(),
		SkillID: skillID.String(),
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	if session.Status != string(domain.LessonSessionStatusActive) {
		t.Fatalf("expected active session, got %+v", session)
	}
	if got := len(repo.queues[mustSessionID(t, session.ID)]); got != defaultSessionLimit {
		t.Fatalf("expected default queue limit %d, got %d", defaultSessionLimit, got)
	}
}

func TestStartSessionAppliesLearnerSelectionRules(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	challenges := []cd.Challenge{
		challengeForSession(skillID, cd.ChallengeTypeMapPoint, 1, `[]`, `{}`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeTheory, 2, `[]`, `[]`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeSingleChoice, 3, `["a"]`, `[]`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeTrueFalse, 4, `["true"]`, `[]`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeFillBlank, 5, `["1905"]`, `[]`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeMatchPairs, 6, `[{"left_id":"l1","right_id":"r1"}]`, `[]`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeMapArea, 7, `{}`, `[]`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeMatchPhotos, 8, `[{"photo_id":"p1","label_id":"l1"}]`, `{"photos":[{"id":"p1","image_url":"https://cdn.test/a.jpg","alt":"A"}],"labels":[{"id":"l1","text":"A"}]}`, cd.ContentStatusPublished),
		challengeForSession(skillID, cd.ChallengeTypeSingleChoice, 9, `["a"]`, `[]`, cd.ContentStatusDraft),
		challengeWithTags(skillID, cd.ChallengeTypeSingleChoice, 10, []string{"placeholder"}),
		challengeWithTags(skillID, cd.ChallengeTypeTrueFalse, 11, []string{"needs_review"}),
	}
	repo := newLearningMemoryRepo()
	content := &contentMemoryRepo{challenges: challenges}
	uc := NewServiceFromRepository(repo, content)

	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID:  userID.String(),
		SkillID: skillID.String(),
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	queue := repo.queues[mustSessionID(t, session.ID)]
	if len(queue) > maxSessionLimit {
		t.Fatalf("expected max %d challenges, got %d", maxSessionLimit, len(queue))
	}
	first := content.challengeByID(queue[0].ChallengeID)
	if first.Type != cd.ChallengeTypeTheory {
		t.Fatalf("expected theory first, got %s", first.Type)
	}
	interactive := 0
	for _, item := range queue {
		challenge := content.challengeByID(item.ChallengeID)
		if challenge.Status != cd.ContentStatusPublished {
			t.Fatalf("non-published challenge selected: %+v", challenge)
		}
		if hasAnyTag(challenge.Tags, "placeholder", "needs_review") {
			t.Fatalf("placeholder/review challenge selected: %+v", challenge)
		}
		switch challenge.Type {
		case cd.ChallengeTypeMapPoint, cd.ChallengeTypeMapArea, cd.ChallengeTypeMatchPhotos:
			interactive++
		}
	}
	if interactive > 1 {
		t.Fatalf("expected at most one interactive challenge, got %d", interactive)
	}
}

func TestStartSessionExcludesUnsafeMatchPhotos(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	unsafePhotos := challengeForSession(skillID, cd.ChallengeTypeMatchPhotos, 1, `[{"photo_id":"p1","label_id":"l1"}]`, `{"photos":[{"id":"p1","image_url":"","alt":"A"}],"labels":[{"id":"l1","text":"A"}]}`, cd.ContentStatusPublished)
	theory := challengeForSession(skillID, cd.ChallengeTypeTheory, 2, `[]`, `[]`, cd.ContentStatusPublished)
	choice := challengeForSession(skillID, cd.ChallengeTypeSingleChoice, 3, `["a"]`, `[]`, cd.ContentStatusPublished)
	repo := newLearningMemoryRepo()
	content := &contentMemoryRepo{challenges: []cd.Challenge{unsafePhotos, theory, choice}}
	uc := NewServiceFromRepository(repo, content)

	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID:  userID.String(),
		SkillID: skillID.String(),
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	for _, item := range repo.queues[mustSessionID(t, session.ID)] {
		if content.challengeByID(item.ChallengeID).Type == cd.ChallengeTypeMatchPhotos {
			t.Fatalf("unsafe match_photos should not be selected")
		}
	}
}

func TestStartSessionUsesChallengePicker(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	challenges := makeChallenges(skillID, 3)
	repo := newLearningMemoryRepo()
	content := &contentMemoryRepo{challenges: challenges}
	picker := &memoryPicker{ids: []domain.ChallengeID{domain.ChallengeID(challenges[2].ID), domain.ChallengeID(challenges[0].ID)}}
	uc := NewService(Dependencies{Sessions: repo, Queue: repo, Attempts: repo, Content: content, Picker: picker})

	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{UserID: userID.String(), SkillID: skillID.String(), Limit: 2})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	queue := repo.queues[mustSessionID(t, session.ID)]
	if len(queue) != 2 || queue[0].ChallengeID != domain.ChallengeID(challenges[2].ID) || queue[1].ChallengeID != domain.ChallengeID(challenges[0].ID) {
		t.Fatalf("expected picker order, got %+v", queue)
	}
}

func TestStartSessionFiltersChallengePickerOutput(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	safe := challengeForSession(skillID, cd.ChallengeTypeSingleChoice, 1, `["a"]`, `[]`, cd.ContentStatusPublished)
	placeholder := challengeWithTags(skillID, cd.ChallengeTypeMapPoint, 2, []string{"placeholder"})
	unsafePhotos := challengeForSession(skillID, cd.ChallengeTypeMatchPhotos, 3, `[{"photo_id":"p1","label_id":"l1"}]`, `{"photos":[{"id":"p1","image_url":"","alt":"A"}],"labels":[{"id":"l1","text":"A"}]}`, cd.ContentStatusPublished)
	content := &contentMemoryRepo{challenges: []cd.Challenge{placeholder, safe, unsafePhotos}}
	repo := newLearningMemoryRepo()
	picker := &memoryPicker{ids: []domain.ChallengeID{
		domain.ChallengeID(placeholder.ID),
		domain.ChallengeID(safe.ID),
		domain.ChallengeID(unsafePhotos.ID),
	}}
	uc := NewService(Dependencies{Sessions: repo, Queue: repo, Attempts: repo, Content: content, Picker: picker})

	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID: userID.String(), SkillID: skillID.String(), Limit: 3,
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	queue := repo.queues[mustSessionID(t, session.ID)]
	if len(queue) != 1 || queue[0].ChallengeID != domain.ChallengeID(safe.ID) {
		t.Fatalf("expected only safe picker challenge, got %+v", queue)
	}
}

func TestSubmitAnswerSavesAttemptAndAdvances(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	content := &contentMemoryRepo{challenges: makeChallenges(skillID, 2)}
	repo := newLearningMemoryRepo()
	uc := NewServiceFromRepository(repo, content)

	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID:  userID.String(),
		SkillID: skillID.String(),
		Limit:   2,
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	current, err := uc.GetCurrentChallenge(context.Background(), learning_api.SessionInput{UserID: userID.String(), SessionID: session.ID})
	if err != nil {
		t.Fatalf("get current failed: %v", err)
	}

	result, err := uc.SubmitAnswer(context.Background(), learning_api.SubmitAnswerInput{
		UserID:     userID.String(),
		SessionID:  session.ID,
		UserAnswer: json.RawMessage(`"a"`),
	})
	if err != nil {
		t.Fatalf("submit answer failed: %v", err)
	}
	if !result.IsCorrect || !result.HasNext {
		t.Fatalf("unexpected submit result: %+v", result)
	}
	sessionID := mustSessionID(t, session.ID)
	if len(repo.attempts[sessionID]) != 1 {
		t.Fatalf("expected one attempt, got %+v", repo.attempts[sessionID])
	}
	if repo.queueByID[mustSessionChallengeID(t, current.SessionChallengeID)].Status != domain.SessionChallengeStatusAnswered {
		t.Fatalf("expected current challenge to be answered")
	}
}

func TestWrongAnswerIsRepeatedBeforeSessionCanFinish(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	content := &contentMemoryRepo{challenges: makeChallenges(skillID, 1)}
	repo := newLearningMemoryRepo()
	uc := NewServiceFromRepository(repo, content)

	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID:  userID.String(),
		SkillID: skillID.String(),
		Limit:   1,
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	wrong, err := uc.SubmitAnswer(context.Background(), learning_api.SubmitAnswerInput{
		UserID:     userID.String(),
		SessionID:  session.ID,
		UserAnswer: json.RawMessage(`"b"`),
	})
	if err != nil {
		t.Fatalf("submit wrong failed: %v", err)
	}
	if wrong.IsCorrect || !wrong.HasNext {
		t.Fatalf("expected wrong answer to repeat, got %+v", wrong)
	}
	sessionID := mustSessionID(t, session.ID)
	if len(repo.queues[sessionID]) != 2 {
		t.Fatalf("expected repeated queue item, got %+v", repo.queues[sessionID])
	}
	if _, err := uc.FinishSession(context.Background(), learning_api.SessionInput{UserID: userID.String(), SessionID: session.ID}); !errors.Is(err, learning_api.ErrInvalidInput) {
		t.Fatalf("expected early finish to fail while retry is pending, got %v", err)
	}
	current, err := uc.GetCurrentChallenge(context.Background(), learning_api.SessionInput{UserID: userID.String(), SessionID: session.ID})
	if err != nil {
		t.Fatalf("get repeated current failed: %v", err)
	}
	if current.Challenge.ID != uuid.UUID(content.challenges[0].ID).String() {
		t.Fatalf("expected same challenge repeated, got %+v", current)
	}
	correct, err := uc.SubmitAnswer(context.Background(), learning_api.SubmitAnswerInput{
		UserID:     userID.String(),
		SessionID:  session.ID,
		UserAnswer: json.RawMessage(`"a"`),
	})
	if err != nil {
		t.Fatalf("submit correct failed: %v", err)
	}
	if !correct.IsCorrect || correct.HasNext {
		t.Fatalf("expected queue drained after retry, got %+v", correct)
	}
	result, err := uc.FinishSession(context.Background(), learning_api.SessionInput{UserID: userID.String(), SessionID: session.ID})
	if err != nil {
		t.Fatalf("finish failed: %v", err)
	}
	if result.Total != 2 || result.Correct != 1 || result.Percent != 50 {
		t.Fatalf("expected attempt-based result, got %+v", result)
	}
}

func TestFinishSessionCalculatesResultAndCallsPorts(t *testing.T) {
	userID := uuid.New()
	skillID := uuid.New()
	repo := newLearningMemoryRepo()
	content := &contentMemoryRepo{challenges: makeChallenges(skillID, 1)}
	progress := &recordingPort{}
	gamification := &recordingPort{}
	uc := NewService(Dependencies{
		Sessions:     repo,
		Queue:        repo,
		Attempts:     repo,
		Content:      content,
		Progress:     progress,
		Gamification: gamification,
	})
	session, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID:  userID.String(),
		SkillID: skillID.String(),
		Limit:   1,
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	if _, err := uc.SubmitAnswer(context.Background(), learning_api.SubmitAnswerInput{UserID: userID.String(), SessionID: session.ID, UserAnswer: json.RawMessage(`"a"`)}); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	result, err := uc.FinishSession(context.Background(), learning_api.SessionInput{UserID: userID.String(), SessionID: session.ID})
	if err != nil {
		t.Fatalf("finish failed: %v", err)
	}
	if result.Total != 1 || result.Correct != 1 || result.Percent != 100 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if progress.calls != 1 || gamification.calls != 1 {
		t.Fatalf("expected hooks to be called once, progress=%d gamification=%d", progress.calls, gamification.calls)
	}
	if _, err := uc.FinishSession(context.Background(), learning_api.SessionInput{UserID: userID.String(), SessionID: session.ID}); err != nil {
		t.Fatalf("second finish failed: %v", err)
	}
	if progress.calls != 1 {
		t.Fatalf("progress should not be called twice, got %d calls", progress.calls)
	}
}

type contentMemoryRepo struct {
	challenges []cd.Challenge
}

func (r *contentMemoryRepo) ListPublishedChallenges(_ context.Context, skillID cd.SkillID) ([]cd.Challenge, error) {
	var out []cd.Challenge
	for _, challenge := range r.challenges {
		if challenge.SkillID == skillID && challenge.Status.IsPublished() {
			out = append(out, challenge)
		}
	}
	return out, nil
}

func (r *contentMemoryRepo) GetChallenge(_ context.Context, id cd.ChallengeID) (cd.Challenge, error) {
	for _, challenge := range r.challenges {
		if challenge.ID == id {
			return challenge, nil
		}
	}
	return cd.Challenge{}, cd.ErrNotFound
}

func (r *contentMemoryRepo) challengeByID(id domain.ChallengeID) cd.Challenge {
	for _, challenge := range r.challenges {
		if domain.ChallengeID(challenge.ID) == id {
			return challenge
		}
	}
	return cd.Challenge{}
}

type learningMemoryRepo struct {
	sessions  map[domain.LessonSessionID]*domain.LessonSession
	queues    map[domain.LessonSessionID][]domain.LessonSessionChallenge
	queueByID map[domain.LessonSessionChallengeID]*domain.LessonSessionChallenge
	attempts  map[domain.LessonSessionID][]domain.ChallengeAttempt
}

func newLearningMemoryRepo() *learningMemoryRepo {
	return &learningMemoryRepo{
		sessions:  map[domain.LessonSessionID]*domain.LessonSession{},
		queues:    map[domain.LessonSessionID][]domain.LessonSessionChallenge{},
		queueByID: map[domain.LessonSessionChallengeID]*domain.LessonSessionChallenge{},
		attempts:  map[domain.LessonSessionID][]domain.ChallengeAttempt{},
	}
}

func (r *learningMemoryRepo) CreateSession(_ context.Context, session *domain.LessonSession) error {
	cp := *session
	r.sessions[session.ID] = &cp
	return nil
}

func (r *learningMemoryRepo) GetSession(_ context.Context, id domain.LessonSessionID) (*domain.LessonSession, error) {
	session, ok := r.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *session
	return &cp, nil
}

func (r *learningMemoryRepo) UpdateSession(_ context.Context, session *domain.LessonSession) error {
	if _, ok := r.sessions[session.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *session
	r.sessions[session.ID] = &cp
	return nil
}

func (r *learningMemoryRepo) CreateMany(_ context.Context, challenges []domain.LessonSessionChallenge) error {
	for _, challenge := range challenges {
		cp := challenge
		r.queues[challenge.SessionID] = append(r.queues[challenge.SessionID], challenge)
		r.queueByID[challenge.ID] = &cp
	}
	return nil
}

func (r *learningMemoryRepo) Append(_ context.Context, challenge domain.LessonSessionChallenge) error {
	queue := r.queues[challenge.SessionID]
	challenge.Position = len(queue) + 1
	cp := challenge
	r.queues[challenge.SessionID] = append(queue, challenge)
	r.queueByID[challenge.ID] = &cp
	return nil
}

func (r *learningMemoryRepo) GetCurrentPending(_ context.Context, sessionID domain.LessonSessionID) (*domain.LessonSessionChallenge, error) {
	for _, challenge := range r.queues[sessionID] {
		stored := r.queueByID[challenge.ID]
		if stored.Status == domain.SessionChallengeStatusPending {
			cp := *stored
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *learningMemoryRepo) MarkAnswered(_ context.Context, id domain.LessonSessionChallengeID) error {
	challenge, ok := r.queueByID[id]
	if !ok {
		return domain.ErrNotFound
	}
	challenge.Status = domain.SessionChallengeStatusAnswered
	return nil
}

func (r *learningMemoryRepo) ListBySession(_ context.Context, sessionID domain.LessonSessionID) ([]domain.LessonSessionChallenge, error) {
	out := make([]domain.LessonSessionChallenge, 0, len(r.queues[sessionID]))
	for _, challenge := range r.queues[sessionID] {
		out = append(out, *r.queueByID[challenge.ID])
	}
	return out, nil
}

func (r *learningMemoryRepo) CreateAttempt(_ context.Context, attempt *domain.ChallengeAttempt) error {
	r.attempts[attempt.SessionID] = append(r.attempts[attempt.SessionID], *attempt)
	return nil
}

func (r *learningMemoryRepo) ListAttemptsBySession(_ context.Context, sessionID domain.LessonSessionID) ([]domain.ChallengeAttempt, error) {
	return r.attempts[sessionID], nil
}

type recordingPort struct {
	calls int
}

type memoryPicker struct {
	ids []domain.ChallengeID
}

func (p *memoryPicker) PickChallengesForSession(context.Context, domain.UserID, domain.SkillID, int) ([]domain.ChallengeID, error) {
	return p.ids, nil
}

func (p *recordingPort) RecordSessionResult(context.Context, ProgressInput) error {
	p.calls++
	return nil
}

func (p *recordingPort) RewardSessionCompleted(context.Context, GamificationInput) error {
	p.calls++
	return nil
}

func makeChallenges(skillID uuid.UUID, count int) []cd.Challenge {
	challenges := make([]cd.Challenge, 0, count)
	for i := 0; i < count; i++ {
		challenges = append(challenges, cd.Challenge{
			ID:          cd.ChallengeID(uuid.New()),
			SkillID:     cd.SkillID(skillID),
			Type:        cd.ChallengeTypeSingleChoice,
			Difficulty:  cd.DifficultyEasy,
			Options:     json.RawMessage(`[{"id":"a","text":"A"},{"id":"b","text":"B"}]`),
			Answers:     json.RawMessage(`["a"]`),
			Explanation: "Because",
			Position:    i + 1,
			Status:      cd.ContentStatusPublished,
		})
	}
	return challenges
}

func challengeForSession(skillID uuid.UUID, challengeType cd.ChallengeType, position int, answers string, options string, status cd.ContentStatus) cd.Challenge {
	if options == "" {
		options = `[]`
	}
	return cd.Challenge{
		ID:          cd.ChallengeID(uuid.New()),
		SkillID:     cd.SkillID(skillID),
		Type:        challengeType,
		Difficulty:  cd.DifficultyEasy,
		Tags:        json.RawMessage(`[]`),
		Options:     json.RawMessage(options),
		Answers:     json.RawMessage(answers),
		Explanation: "Because",
		Position:    position,
		Status:      status,
	}
}

func challengeWithTags(skillID uuid.UUID, challengeType cd.ChallengeType, position int, tags []string) cd.Challenge {
	rawTags, _ := json.Marshal(tags)
	return cd.Challenge{
		ID:          cd.ChallengeID(uuid.New()),
		SkillID:     cd.SkillID(skillID),
		Type:        challengeType,
		Difficulty:  cd.DifficultyEasy,
		Tags:        json.RawMessage(rawTags),
		Options:     json.RawMessage(`[]`),
		Answers:     json.RawMessage(`["a"]`),
		Explanation: "Because",
		Position:    position,
		Status:      cd.ContentStatusPublished,
	}
}

func mustSessionID(t *testing.T, raw string) domain.LessonSessionID {
	t.Helper()
	id, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("invalid session id: %v", err)
	}
	return domain.LessonSessionID(id)
}

func mustSessionChallengeID(t *testing.T, raw string) domain.LessonSessionChallengeID {
	t.Helper()
	id, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("invalid session challenge id: %v", err)
	}
	return domain.LessonSessionChallengeID(id)
}

func TestStartSessionReturnsNoChallenges(t *testing.T) {
	uc := NewServiceFromRepository(newLearningMemoryRepo(), &contentMemoryRepo{})
	_, err := uc.StartSession(context.Background(), learning_api.StartSessionInput{
		UserID:  uuid.New().String(),
		SkillID: uuid.New().String(),
	})
	if !errors.Is(err, learning_api.ErrNoChallenges) {
		t.Fatalf("expected no challenges error, got %v", err)
	}
}
