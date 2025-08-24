package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RepositoryTestSuite struct {
	suite.Suite
	repo *Repository
	db   *TestDB
}

func (suite *RepositoryTestSuite) SetupSuite() {
	suite.db = NewTestDB(suite.T())
	suite.repo = NewRepository(suite.db.Pool)
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *RepositoryTestSuite) TestRepositoryInitialization() {
	// Test that all repositories are properly initialized
	assert.NotNil(suite.T(), suite.repo.User)
	assert.NotNil(suite.T(), suite.repo.Session)
	assert.NotNil(suite.T(), suite.repo.Message)
	assert.NotNil(suite.T(), suite.repo.Logbook)
}

func (suite *RepositoryTestSuite) TestRepositoryTypes() {
	// Test that repositories are of correct types
	assert.IsType(suite.T(), &UserRepository{}, suite.repo.User)
	assert.IsType(suite.T(), &SessionRepository{}, suite.repo.Session)
	assert.IsType(suite.T(), &MessageRepository{}, suite.repo.Message)
	assert.IsType(suite.T(), &LogbookRepository{}, suite.repo.Logbook)
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}