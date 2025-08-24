package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"bodda/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogbookRepository is a mock implementation of LogbookRepositoryInterface
type MockLogbookRepository struct {
	mock.Mock
}

func (m *MockLogbookRepository) Create(ctx context.Context, logbook *models.AthleteLogbook) error {
	args := m.Called(ctx, logbook)
	// Simulate database behavior by setting ID and timestamp
	if args.Error(0) == nil {
		logbook.ID = "test-id"
		logbook.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockLogbookRepository) GetByID(ctx context.Context, id string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookRepository) GetByUserID(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookRepository) Update(ctx context.Context, logbook *models.AthleteLogbook) error {
	args := m.Called(ctx, logbook)
	// Simulate database behavior by updating timestamp
	if args.Error(0) == nil {
		logbook.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockLogbookRepository) Upsert(ctx context.Context, logbook *models.AthleteLogbook) error {
	args := m.Called(ctx, logbook)
	// Simulate database behavior by setting ID and timestamp
	if args.Error(0) == nil {
		if logbook.ID == "" {
			logbook.ID = "test-id"
		}
		logbook.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockLogbookRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNewLogbookService(t *testing.T) {
	mockRepo := &MockLogbookRepository{}
	service := NewLogbookService(mockRepo)
	
	assert.NotNil(t, service)
	assert.IsType(t, &logbookService{}, service)
}

func TestLogbookService_GetLogbook(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockLogbookRepository)
		expectedError  string
		expectedResult *models.AthleteLogbook
	}{
		{
			name:          "empty user ID",
			userID:        "",
			mockSetup:     func(m *MockLogbookRepository) {},
			expectedError: "user ID cannot be empty",
		},
		{
			name:   "logbook not found",
			userID: "user-123",
			mockSetup: func(m *MockLogbookRepository) {
				m.On("GetByUserID", mock.Anything, "user-123").Return(nil, fmt.Errorf("logbook not found"))
			},
			expectedError: "logbook not found for user user-123",
		},
		{
			name:   "successful retrieval",
			userID: "user-123",
			mockSetup: func(m *MockLogbookRepository) {
				logbook := &models.AthleteLogbook{
					ID:      "logbook-123",
					UserID:  "user-123",
					Content: `{"personal_info":{"name":"Test User"}}`,
				}
				m.On("GetByUserID", mock.Anything, "user-123").Return(logbook, nil)
			},
			expectedResult: &models.AthleteLogbook{
				ID:      "logbook-123",
				UserID:  "user-123",
				Content: `{"personal_info":{"name":"Test User"}}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockLogbookRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewLogbookService(mockRepo)
			result, err := service.GetLogbook(context.Background(), tt.userID)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLogbookService_CreateInitialLogbook(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		stravaProfile  *StravaAthlete
		mockSetup      func(*MockLogbookRepository)
		expectedError  string
		validateResult func(*testing.T, *models.AthleteLogbook)
	}{
		{
			name:          "empty user ID",
			userID:        "",
			stravaProfile: &StravaAthlete{},
			mockSetup:     func(m *MockLogbookRepository) {},
			expectedError: "user ID cannot be empty",
		},
		{
			name:          "nil strava profile",
			userID:        "user-123",
			stravaProfile: nil,
			mockSetup:     func(m *MockLogbookRepository) {},
			expectedError: "strava profile cannot be nil",
		},
		{
			name:   "successful creation",
			userID: "user-123",
			stravaProfile: &StravaAthlete{
				ID:        12345,
				Firstname: "John",
				Lastname:  "Doe",
				Sex:       "M",
				City:      "San Francisco",
				State:     "CA",
				Country:   "USA",
				Weight:    70.5,
				FTP:       250,
				CreatedAt: "2020-01-01T00:00:00Z",
			},
			mockSetup: func(m *MockLogbookRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.AthleteLogbook")).Return(nil)
			},
			validateResult: func(t *testing.T, result *models.AthleteLogbook) {
				assert.Equal(t, "user-123", result.UserID)
				assert.NotEmpty(t, result.Content)
				
				// Validate the string content contains expected information
				assert.Contains(t, result.Content, "John Doe")
				assert.Contains(t, result.Content, "M")
				assert.Contains(t, result.Content, "San Francisco, CA, USA")
				assert.Contains(t, result.Content, "70.5 kg")
				assert.Contains(t, result.Content, "250 watts")
				assert.Contains(t, result.Content, "2020-01-01T00:00:00Z")
				assert.Contains(t, result.Content, "Initial logbook created")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockLogbookRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewLogbookService(mockRepo)
			result, err := service.CreateInitialLogbook(context.Background(), tt.userID, tt.stravaProfile)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLogbookService_UpdateLogbook(t *testing.T) {
	existingContent := `Athlete Profile:
Name: John Doe
Age: 30
Gender: M
FTP: 250 watts

Training Notes:
- Initial logbook created`

	tests := []struct {
		name           string
		userID         string
		content        string
		mockSetup      func(*MockLogbookRepository)
		expectedError  string
		validateResult func(*testing.T, *models.AthleteLogbook)
	}{
		{
			name:          "empty user ID",
			userID:        "",
			content:       "some content",
			mockSetup:     func(m *MockLogbookRepository) {},
			expectedError: "user ID cannot be empty",
		},
		{
			name:          "empty content",
			userID:        "user-123",
			content:       "",
			mockSetup:     func(m *MockLogbookRepository) {},
			expectedError: "content cannot be empty",
		},
		{
			name:    "logbook not found",
			userID:  "user-123",
			content: "Updated content",
			mockSetup: func(m *MockLogbookRepository) {
				m.On("GetByUserID", mock.Anything, "user-123").Return(nil, fmt.Errorf("logbook not found"))
			},
			expectedError: "logbook not found for user user-123",
		},
		{
			name:    "successful update",
			userID:  "user-123",
			content: "Updated athlete profile with new training insights and goals",
			mockSetup: func(m *MockLogbookRepository) {
				existingLogbook := &models.AthleteLogbook{
					ID:      "logbook-123",
					UserID:  "user-123",
					Content: existingContent,
				}
				m.On("GetByUserID", mock.Anything, "user-123").Return(existingLogbook, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.AthleteLogbook")).Return(nil)
			},
			validateResult: func(t *testing.T, result *models.AthleteLogbook) {
				assert.Equal(t, "user-123", result.UserID)
				assert.Equal(t, "Updated athlete profile with new training insights and goals", result.Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockLogbookRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewLogbookService(mockRepo)
			result, err := service.UpdateLogbook(context.Background(), tt.userID, tt.content)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLogbookService_UpsertLogbook(t *testing.T) {
	validContent := `Athlete Profile:
Name: Test User
Training insights and goals go here.`
	
	tests := []struct {
		name          string
		userID        string
		content       string
		mockSetup     func(*MockLogbookRepository)
		expectedError string
	}{
		{
			name:          "empty user ID",
			userID:        "",
			content:       validContent,
			mockSetup:     func(m *MockLogbookRepository) {},
			expectedError: "user ID cannot be empty",
		},
		{
			name:          "empty content",
			userID:        "user-123",
			content:       "",
			mockSetup:     func(m *MockLogbookRepository) {},
			expectedError: "content cannot be empty",
		},
		{
			name:    "successful upsert",
			userID:  "user-123",
			content: validContent,
			mockSetup: func(m *MockLogbookRepository) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*models.AthleteLogbook")).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockLogbookRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewLogbookService(mockRepo)
			result, err := service.UpsertLogbook(context.Background(), tt.userID, tt.content)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.UserID)
				assert.Equal(t, tt.content, result.Content)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}



func TestHelperFunctions(t *testing.T) {
	t.Run("formatLocation", func(t *testing.T) {
		tests := []struct {
			city, state, country string
			expected             string
		}{
			{"San Francisco", "CA", "USA", "San Francisco, CA, USA"},
			{"", "CA", "USA", "CA, USA"},
			{"San Francisco", "", "USA", "San Francisco, USA"},
			{"San Francisco", "CA", "", "San Francisco, CA"},
			{"", "", "", ""},
		}

		for _, tt := range tests {
			result := formatLocation(tt.city, tt.state, tt.country)
			assert.Equal(t, tt.expected, result)
		}
	})
}