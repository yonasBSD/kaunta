package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seuros/kaunta/internal/database"
)

func TestNewEventPayloadConvertsUUIDsToStrings(t *testing.T) {
	websiteID := uuid.New()
	sessionID := uuid.New()
	visitID := uuid.New()
	createdAt := time.Now()

	payload := NewEventPayload("visit", websiteID, sessionID, visitID, "/page", "Title", createdAt)

	require.Equal(t, "visit", payload.Type)
	require.Equal(t, websiteID.String(), payload.WebsiteID)
	require.Equal(t, sessionID.String(), payload.SessionID)
	require.Equal(t, visitID.String(), payload.VisitID)
	require.Equal(t, "/page", payload.Path)
	require.Equal(t, "Title", payload.Title)
	require.WithinDuration(t, createdAt, payload.CreatedAt, time.Millisecond)
}

func TestNotifyEventPublishesPayload(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = mockDB.Close() })

	originalDB := database.DB
	database.DB = mockDB
	t.Cleanup(func() { database.DB = originalDB })

	payload := EventPayload{
		Type:      "visit",
		WebsiteID: uuid.NewString(),
		SessionID: uuid.NewString(),
		VisitID:   uuid.NewString(),
		Path:      "/test",
		Title:     "Test",
		CreatedAt: time.Now(),
	}

	bytes, err := json.Marshal(payload)
	require.NoError(t, err)

	mock.ExpectExec("SELECT pg_notify").
		WithArgs(ChannelName, string(bytes)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	NotifyEvent(context.Background(), payload)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyEventHandlesExecError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = mockDB.Close() })

	originalDB := database.DB
	database.DB = mockDB
	t.Cleanup(func() { database.DB = originalDB })

	payload := EventPayload{
		Type:      "visit",
		WebsiteID: uuid.NewString(),
		SessionID: uuid.NewString(),
		VisitID:   uuid.NewString(),
		CreatedAt: time.Now(),
	}

	bytes, err := json.Marshal(payload)
	require.NoError(t, err)

	mock.ExpectExec("SELECT pg_notify").
		WithArgs(ChannelName, string(bytes)).
		WillReturnError(assert.AnError)

	NotifyEvent(context.Background(), payload)

	require.NoError(t, mock.ExpectationsWereMet())
}
