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
