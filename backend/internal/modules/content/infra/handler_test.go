package content_infra_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	content_api "github.com/SIniutin/history-app-backend/internal/modules/content/api"
	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	content_infra "github.com/SIniutin/history-app-backend/internal/modules/content/infra"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	users_security "github.com/SIniutin/history-app-backend/internal/modules/users/security"
	"github.com/google/uuid"
)

func TestListCourses(t *testing.T) {
	handler := content_infra.NewHandler(content_infra.Dependencies{
		Courses: fakeContent{},
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	handler.Routes().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var courses []content_api.Course
	if err := json.NewDecoder(res.Body).Decode(&courses); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != "История России" {
		t.Fatalf("unexpected courses: %+v", courses)
	}
}

func TestNestedEndpointRejectsInvalidUUID(t *testing.T) {
	handler := content_infra.NewHandler(content_infra.Dependencies{
		Sections: fakeContent{sectionErr: content_api.ErrInvalidInput},
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/courses/not-a-uuid/sections", nil)
	handler.Routes().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestChallengesResponseDoesNotIncludeAnswers(t *testing.T) {
	handler := content_infra.NewHandler(content_infra.Dependencies{
		Challenges: fakeContent{},
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/skills/55555555-5555-4555-8555-555555555555/challenges", nil)
	handler.Routes().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "answers") {
		t.Fatalf("response must not expose answers: %s", res.Body.String())
	}
}

func TestAuthoringRoleChecks(t *testing.T) {
	auth := users_infra.NewAuthMiddleware([]byte("access-secret"), "history-app-backend", "history-app-users")
	content := fakeContent{}
	handler := content_infra.NewHandler(content_infra.Dependencies{
		Courses:    content,
		Sections:   content,
		Units:      content,
		Skills:     content,
		Challenges: content,
		Auth:       auth,
	})

	editorToken := accessToken(t, ud.RoleContentEditor)
	reviewerToken := accessToken(t, ud.RoleContentReviewer)

	editorPublish := doJSON(handler.Routes(), http.MethodPost, "/editor/content/courses/11111111-1111-4111-8111-111111111111/publish", nil, editorToken)
	if editorPublish.Code != http.StatusForbidden {
		t.Fatalf("editor publish status=%d body=%s", editorPublish.Code, editorPublish.Body.String())
	}

	reviewerCreate := doJSON(handler.Routes(), http.MethodPost, "/editor/content/courses", map[string]string{
		"source_lang": "ru",
		"target_lang": "ru",
		"title":       "История",
	}, reviewerToken)
	if reviewerCreate.Code != http.StatusForbidden {
		t.Fatalf("reviewer create status=%d body=%s", reviewerCreate.Code, reviewerCreate.Body.String())
	}
}

func accessToken(t *testing.T, role ud.Role) string {
	t.Helper()
	tokens := users_security.NewTokenService(users_security.TokenConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    time.Hour,
		Issuer:        "history-app-backend",
		Audience:      "history-app-users",
	})
	pair, _, err := tokens.NewPair(ud.UserID(uuid.New()), role)
	if err != nil {
		t.Fatalf("new token pair: %v", err)
	}
	return pair.AccessToken
}

func doJSON(handler http.Handler, method, path string, body any, accessToken string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == nil {
		reader = strings.NewReader("")
	} else {
		raw, _ := json.Marshal(body)
		reader = strings.NewReader(string(raw))
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

type fakeContent struct {
	sectionErr error
}

func (f fakeContent) ListPublishedCourses(context.Context) ([]cd.Course, error) {
	return []cd.Course{{ID: cd.CourseID(uuid.MustParse("11111111-1111-4111-8111-111111111111")), Title: "История России"}}, nil
}

func (f fakeContent) ListAllCourses(context.Context) ([]cd.Course, error) { return nil, nil }
func (f fakeContent) CreateCourse(context.Context, content_api.CourseWriteInput) (cd.Course, error) {
	return cd.Course{}, nil
}
func (f fakeContent) UpdateCourse(context.Context, content_api.CourseWriteInput) (cd.Course, error) {
	return cd.Course{}, nil
}
func (f fakeContent) ListPublishedSections(context.Context, content_api.ListSectionsInput) ([]cd.Section, error) {
	return nil, f.sectionErr
}
func (f fakeContent) ListAllSections(context.Context, content_api.ListSectionsInput) ([]cd.Section, error) {
	return nil, nil
}
func (f fakeContent) CreateSection(context.Context, content_api.SectionWriteInput) (cd.Section, error) {
	return cd.Section{}, nil
}
func (f fakeContent) UpdateSection(context.Context, content_api.SectionWriteInput) (cd.Section, error) {
	return cd.Section{}, nil
}
func (f fakeContent) ListPublishedUnits(context.Context, content_api.ListUnitsInput) ([]cd.Unit, error) {
	return nil, nil
}
func (f fakeContent) ListAllUnits(context.Context, content_api.ListUnitsInput) ([]cd.Unit, error) {
	return nil, nil
}
func (f fakeContent) CreateUnit(context.Context, content_api.UnitWriteInput) (cd.Unit, error) {
	return cd.Unit{}, nil
}
func (f fakeContent) UpdateUnit(context.Context, content_api.UnitWriteInput) (cd.Unit, error) {
	return cd.Unit{}, nil
}
func (f fakeContent) ListPublishedSkills(context.Context, content_api.ListSkillsInput) ([]cd.Skill, error) {
	return nil, nil
}
func (f fakeContent) ListAllSkills(context.Context, content_api.ListSkillsInput) ([]cd.Skill, error) {
	return nil, nil
}
func (f fakeContent) CreateSkill(context.Context, content_api.SkillWriteInput) (cd.Skill, error) {
	return cd.Skill{}, nil
}
func (f fakeContent) UpdateSkill(context.Context, content_api.SkillWriteInput) (cd.Skill, error) {
	return cd.Skill{}, nil
}
func (f fakeContent) ListPublishedChallenges(context.Context, content_api.ListChallengesInput) ([]cd.Challenge, error) {
	return []cd.Challenge{
		{
			ID:          cd.ChallengeID(uuid.MustParse("66666666-6666-4666-8666-666666666662")),
			SkillID:     cd.SkillID(uuid.MustParse("55555555-5555-4555-8555-555555555555")),
			Type:        cd.ChallengeTypeSingleChoice,
			Difficulty:  cd.DifficultyEasy,
			Tags:        []byte(`["seed"]`),
			Prompt:      "В каком году Николай II вступил на престол?",
			Options:     []byte(`["1894","1905"]`),
			Explanation: "Николай II вступил на престол в 1894 году.",
		},
	}, nil
}
func (f fakeContent) ListAllChallenges(context.Context, content_api.ListChallengesInput) ([]cd.Challenge, error) {
	return nil, nil
}
func (f fakeContent) CreateChallenge(context.Context, content_api.ChallengeWriteInput) (cd.Challenge, error) {
	return cd.Challenge{}, nil
}
func (f fakeContent) UpdateChallenge(context.Context, content_api.ChallengeWriteInput) (cd.Challenge, error) {
	return cd.Challenge{}, nil
}
func (f fakeContent) Publish(context.Context, content_api.StatusTransitionInput) error { return nil }
func (f fakeContent) Archive(context.Context, content_api.StatusTransitionInput) error { return nil }
