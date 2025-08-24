package database

import (
	"context"
	"testing"
	"time"

	"bodda/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	repo *UserRepository
	db   *TestDB
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	suite.db = NewTestDB(suite.T())
	suite.repo = NewUserRepository(suite.db.Pool)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.db.CleanTables()
}

func (suite *UserRepositoryTestSuite) TestCreateUser() {
	user := &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}

	err := suite.repo.Create(context.Background(), user)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), user.ID)
	assert.NotZero(suite.T(), user.CreatedAt)
	assert.NotZero(suite.T(), user.UpdatedAt)
}

func (suite *UserRepositoryTestSuite) TestGetUserByID() {
	// Create a user first
	user := &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := suite.repo.Create(context.Background(), user)
	assert.NoError(suite.T(), err)

	// Get the user by ID
	retrievedUser, err := suite.repo.GetByID(context.Background(), user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrievedUser.ID)
	assert.Equal(suite.T(), user.StravaID, retrievedUser.StravaID)
	assert.Equal(suite.T(), user.FirstName, retrievedUser.FirstName)
	assert.Equal(suite.T(), user.LastName, retrievedUser.LastName)
}

func (suite *UserRepositoryTestSuite) TestGetUserByStravaID() {
	// Create a user first
	user := &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := suite.repo.Create(context.Background(), user)
	assert.NoError(suite.T(), err)

	// Get the user by Strava ID
	retrievedUser, err := suite.repo.GetByStravaID(context.Background(), user.StravaID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrievedUser.ID)
	assert.Equal(suite.T(), user.StravaID, retrievedUser.StravaID)
	assert.Equal(suite.T(), user.FirstName, retrievedUser.FirstName)
}

func (suite *UserRepositoryTestSuite) TestUpdateUser() {
	// Create a user first
	user := &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := suite.repo.Create(context.Background(), user)
	assert.NoError(suite.T(), err)

	// Update the user
	user.FirstName = "Jane"
	user.AccessToken = "new_access_token"
	originalUpdatedAt := user.UpdatedAt

	err = suite.repo.Update(context.Background(), user)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), user.UpdatedAt.After(originalUpdatedAt))

	// Verify the update
	retrievedUser, err := suite.repo.GetByID(context.Background(), user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Jane", retrievedUser.FirstName)
	assert.Equal(suite.T(), "new_access_token", retrievedUser.AccessToken)
}

func (suite *UserRepositoryTestSuite) TestDeleteUser() {
	// Create a user first
	user := &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := suite.repo.Create(context.Background(), user)
	assert.NoError(suite.T(), err)

	// Delete the user
	err = suite.repo.Delete(context.Background(), user.ID)
	assert.NoError(suite.T(), err)

	// Verify the user is deleted
	_, err = suite.repo.GetByID(context.Background(), user.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestGetNonExistentUser() {
	_, err := suite.repo.GetByID(context.Background(), "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}