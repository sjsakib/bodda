package database

import (
	"context"
	"testing"
	"time"

	"bodda/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LogbookRepositoryTestSuite struct {
	suite.Suite
	repo     *LogbookRepository
	userRepo *UserRepository
	db       *TestDB
	testUser *models.User
}

func (suite *LogbookRepositoryTestSuite) SetupSuite() {
	suite.db = NewTestDB(suite.T())
	suite.repo = NewLogbookRepository(suite.db.Pool)
	suite.userRepo = NewUserRepository(suite.db.Pool)
}

func (suite *LogbookRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *LogbookRepositoryTestSuite) SetupTest() {
	suite.db.CleanTables()
	
	// Create a test user
	suite.testUser = &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := suite.userRepo.Create(context.Background(), suite.testUser)
	assert.NoError(suite.T(), err)
}

func (suite *LogbookRepositoryTestSuite) TestCreateLogbook() {
	logbook := &models.AthleteLogbook{
		UserID: suite.testUser.ID,
		Content: `Athlete Profile:
Name: John Doe
Training insights and goals will be updated here as we learn more about the athlete.`,
	}

	err := suite.repo.Create(context.Background(), logbook)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), logbook.ID)
	assert.NotZero(suite.T(), logbook.UpdatedAt)
}

func (suite *LogbookRepositoryTestSuite) TestGetLogbookByID() {
	// Create a logbook first
	logbook := &models.AthleteLogbook{
		UserID: suite.testUser.ID,
		Content: `Athlete Profile:
Name: Test User
Training data and insights go here.`,
	}
	err := suite.repo.Create(context.Background(), logbook)
	assert.NoError(suite.T(), err)

	// Get the logbook by ID
	retrievedLogbook, err := suite.repo.GetByID(context.Background(), logbook.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), logbook.ID, retrievedLogbook.ID)
	assert.Equal(suite.T(), logbook.UserID, retrievedLogbook.UserID)
	assert.Equal(suite.T(), logbook.Content, retrievedLogbook.Content)
}

func (suite *LogbookRepositoryTestSuite) TestGetLogbookByUserID() {
	// Create a logbook first
	logbook := &models.AthleteLogbook{
		UserID: suite.testUser.ID,
		Content: `Athlete Profile:
Name: Test User
Training data and insights go here.`,
	}
	err := suite.repo.Create(context.Background(), logbook)
	assert.NoError(suite.T(), err)

	// Get the logbook by user ID
	retrievedLogbook, err := suite.repo.GetByUserID(context.Background(), suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), logbook.ID, retrievedLogbook.ID)
	assert.Equal(suite.T(), logbook.UserID, retrievedLogbook.UserID)
	assert.Equal(suite.T(), logbook.Content, retrievedLogbook.Content)
}

func (suite *LogbookRepositoryTestSuite) TestUpdateLogbook() {
	// Create a logbook first
	logbook := &models.AthleteLogbook{
		UserID: suite.testUser.ID,
		Content: `Athlete Profile:
Name: Test User
Original training data and insights.`,
	}
	err := suite.repo.Create(context.Background(), logbook)
	assert.NoError(suite.T(), err)

	// Update the logbook
	logbook.Content = `Athlete Profile:
Name: Test User
Updated training data and insights with new goals and observations.`
	originalUpdatedAt := logbook.UpdatedAt

	err = suite.repo.Update(context.Background(), logbook)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), logbook.UpdatedAt.After(originalUpdatedAt))

	// Verify the update
	retrievedLogbook, err := suite.repo.GetByID(context.Background(), logbook.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `Athlete Profile:
Name: Test User
Updated training data and insights with new goals and observations.`, retrievedLogbook.Content)
}

func (suite *LogbookRepositoryTestSuite) TestUpsertLogbook() {
	// Test insert (logbook doesn't exist)
	logbook := &models.AthleteLogbook{
		UserID: suite.testUser.ID,
		Content: `Athlete Profile:
Name: Test User
Initial training data and insights.`,
	}

	err := suite.repo.Upsert(context.Background(), logbook)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), logbook.ID)
	assert.NotZero(suite.T(), logbook.UpdatedAt)

	// Verify it was created
	retrievedLogbook, err := suite.repo.GetByUserID(context.Background(), suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `Athlete Profile:
Name: Test User
Initial training data and insights.`, retrievedLogbook.Content)

	// Test update (logbook exists)
	logbook.Content = `Athlete Profile:
Name: Test User
Updated training data and insights with new goals.`
	originalUpdatedAt := logbook.UpdatedAt

	err = suite.repo.Upsert(context.Background(), logbook)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), logbook.UpdatedAt.After(originalUpdatedAt))

	// Verify it was updated, not duplicated
	retrievedLogbook, err = suite.repo.GetByUserID(context.Background(), suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `Athlete Profile:
Name: Test User
Updated training data and insights with new goals.`, retrievedLogbook.Content)
	assert.Equal(suite.T(), logbook.ID, retrievedLogbook.ID) // Same ID, not a new record
}

func (suite *LogbookRepositoryTestSuite) TestDeleteLogbook() {
	// Create a logbook first
	logbook := &models.AthleteLogbook{
		UserID: suite.testUser.ID,
		Content: `Athlete Profile:
Name: Test User
Test logbook content for deletion.`,
	}
	err := suite.repo.Create(context.Background(), logbook)
	assert.NoError(suite.T(), err)

	// Delete the logbook
	err = suite.repo.Delete(context.Background(), logbook.ID)
	assert.NoError(suite.T(), err)

	// Verify the logbook is deleted
	_, err = suite.repo.GetByID(context.Background(), logbook.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "logbook not found")
}

func (suite *LogbookRepositoryTestSuite) TestGetNonExistentLogbook() {
	_, err := suite.repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "logbook not found")

	_, err = suite.repo.GetByUserID(context.Background(), "00000000-0000-0000-0000-000000000001")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "logbook not found")
}

func (suite *LogbookRepositoryTestSuite) TestUniqueConstraintOnUserID() {
	// Create first logbook
	logbook1 := &models.AthleteLogbook{
		UserID:  suite.testUser.ID,
		Content: "First logbook",
	}
	err := suite.repo.Create(context.Background(), logbook1)
	assert.NoError(suite.T(), err)

	// Try to create second logbook for same user (should fail due to unique constraint)
	logbook2 := &models.AthleteLogbook{
		UserID:  suite.testUser.ID,
		Content: "Second logbook",
	}
	err = suite.repo.Create(context.Background(), logbook2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate key value violates unique constraint")
}

func TestLogbookRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(LogbookRepositoryTestSuite))
}