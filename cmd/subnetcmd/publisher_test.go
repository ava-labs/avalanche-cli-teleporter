// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package subnetcmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanche-cli/internal/mocks"
	"github.com/ava-labs/avalanche-cli/pkg/application"
	"github.com/ava-labs/avalanche-cli/pkg/config"
	"github.com/ava-labs/avalanche-cli/pkg/constants"
	"github.com/ava-labs/avalanche-cli/pkg/models"
	"github.com/ava-labs/avalanche-cli/pkg/subnet"
	"github.com/ava-labs/avalanche-cli/pkg/ux"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestPublisher(string, string, string) subnet.Publisher {
	mockPub := &mocks.Publisher{}
	mockPub.On("GetRepo").Return(&git.Repository{}, nil)
	mockPub.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return mockPub
}

func TestIsPublished(t *testing.T) {
	assert, _ := setupTestEnv(t)
	defer func() {
		app = nil
	}()

	testSubnet := "testSubnet"

	published, err := isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.False(published)

	baseDir := app.GetBaseDir()
	err = os.Mkdir(filepath.Join(baseDir, testSubnet), constants.DefaultPerms755)
	assert.NoError(err)
	published, err = isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.False(published)

	reposDir := app.GetReposDir()
	err = os.MkdirAll(filepath.Join(reposDir, "dummyRepo", constants.VMDir, testSubnet), constants.DefaultPerms755)
	assert.NoError(err)
	published, err = isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.False(published)

	goodDir1 := filepath.Join(reposDir, "dummyRepo", constants.SubnetDir, testSubnet)
	err = os.MkdirAll(goodDir1, constants.DefaultPerms755)
	assert.NoError(err)
	published, err = isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.False(published)

	_, err = os.Create(filepath.Join(goodDir1, testSubnet))
	assert.NoError(err)
	published, err = isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.True(published)

	goodDir2 := filepath.Join(reposDir, "dummyRepo2", constants.SubnetDir, testSubnet)
	err = os.MkdirAll(goodDir2, constants.DefaultPerms755)
	assert.NoError(err)
	published, err = isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.True(published)
	_, err = os.Create(filepath.Join(goodDir2, "myOtherTestSubnet"))
	assert.NoError(err)
	published, err = isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.True(published)

	_, err = os.Create(filepath.Join(goodDir2, testSubnet))
	assert.NoError(err)
	published, err = isAlreadyPublished(testSubnet)
	assert.NoError(err)
	assert.True(published)
}

// TestPublisher allows unit testing of the **normal** flow for publishing
func TestPublisher(t *testing.T) {
	assert, mockPrompt := setupTestEnv(t)
	defer func() {
		app = nil
	}()

	// capture string for a repo alias...
	mockPrompt.On("CaptureString", mock.Anything).Return("testAlias", nil).Once()
	// then the repo URL...
	mockPrompt.On("CaptureString", mock.Anything).Return("https://localhost:12345", nil).Once()
	// always provide an irrelevant response when empty is allowed...
	mockPrompt.On("CaptureEmpty", mock.Anything, mock.Anything).Return("irrelevant", nil)
	// on the maintainers, return some array
	mockPrompt.On("CaptureListDecision", mockPrompt, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]any{"dummy", "stuff"}, false, nil)
	// finally return a semantic version
	mockPrompt.On("CaptureVersion", mock.Anything).Return("v0.9.99", nil)

	sc := &models.Sidecar{
		VM: models.SubnetEvm,
	}
	err := doPublish(sc, "testSubnet", newTestPublisher)
	assert.NoError(err)
}

func setupTestEnv(t *testing.T) (*assert.Assertions, *mocks.Prompter) {
	assert := assert.New(t)
	testDir := t.TempDir()
	err := os.Mkdir(filepath.Join(testDir, "repos"), 0o755)
	assert.NoError(err)
	ux.NewUserLog(logging.NoLog{}, io.Discard)
	app = &application.Avalanche{}
	mockPrompt := mocks.NewPrompter(t)
	app.Setup(testDir, logging.NoLog{}, config.New(), mockPrompt)

	return assert, mockPrompt
}
