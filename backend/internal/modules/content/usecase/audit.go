package usecase

import (
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
)

func newAudit(actorID domain.UserID, now time.Time) domain.Audit {
	return domain.Audit{CreatedBy: actorID, UpdatedBy: actorID, CreatedAt: now, UpdatedAt: now}
}

func updateAudit(actorID domain.UserID) domain.Audit {
	return domain.Audit{UpdatedBy: actorID, UpdatedAt: time.Now().UTC()}
}
