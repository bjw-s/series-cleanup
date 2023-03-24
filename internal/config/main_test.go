package config

import (
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestTraktConfigTestSuite struct {
	suite.Suite
}

type TestPlexConfigTestSuite struct {
	suite.Suite
}

type TestFolderRulesTestSuite struct {
	suite.Suite
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	validate = validator.New()
	validate.RegisterValidation("valid_folder", validateFolder)
}

func teardown() {
}

func TestTraktConfig(t *testing.T) {
	suite.Run(t, new(TestTraktConfigTestSuite))
}

func (suite *TestTraktConfigTestSuite) TestDisabled() {
	var err error
	var tc = traktConfig{}
	tc.Enabled = false

	err = validate.Struct(tc)
	assert.NoError(suite.T(), err)
}

func (suite *TestTraktConfigTestSuite) TestValidValues() {
	workingDir, _ := os.Getwd()
	var err error
	var tc = traktConfig{}
	tc.Enabled = true
	tc.ClientID = "12345"
	tc.ClientSecret = "12345"
	tc.User = "12345"
	tc.CacheDir = workingDir

	err = validate.Struct(tc)
	assert.NoError(suite.T(), err)
}

func (suite *TestTraktConfigTestSuite) TestRequiredFields() {
	var err error
	var tc = traktConfig{}
	tc.Enabled = true

	requiredFields := []string{
		"traktConfig.CacheDir", "traktConfig.ClientID", "traktConfig.ClientSecret", "traktConfig.User",
	}

	err = validate.Struct(tc)
	if assert.Error(suite.T(), err) {
		assert.IsType(suite.T(), validator.ValidationErrors{}, err)
		validationErrors := err.(validator.ValidationErrors)
		for i, field := range requiredFields {
			assert.Equal(suite.T(), "required_if", validationErrors[i].Tag())
			assert.Equal(suite.T(), field, validationErrors[i].Namespace())
		}
	}
}

func (suite *TestTraktConfigTestSuite) TestInvalidCacheFolder() {
	var err error
	var tc = traktConfig{}
	tc.Enabled = true
	tc.User = "test"
	tc.ClientID = "12345"
	tc.ClientSecret = "12345"
	tc.CacheDir = "/this_does_not_exist"

	err = validate.Struct(tc)
	if assert.Error(suite.T(), err) {
		assert.IsType(suite.T(), validator.ValidationErrors{}, err)
		validationErrors := err.(validator.ValidationErrors)
		assert.Equal(suite.T(), "valid_folder", validationErrors[0].Tag())
		assert.Equal(suite.T(), "traktConfig.CacheDir", validationErrors[0].Namespace())
	}
}

func TestPlexConfig(t *testing.T) {
	suite.Run(t, new(TestPlexConfigTestSuite))
}

func (suite *TestPlexConfigTestSuite) TestDisabled() {
	var err error
	var pc = plexConfig{}
	pc.Enabled = false
	assert.NoError(suite.T(), err)
}

func (suite *TestPlexConfigTestSuite) TestValidValues() {
	var err error
	var pc = plexConfig{}
	pc.Enabled = true
	pc.Token = "12345"
	assert.NoError(suite.T(), err)
}

func (suite *TestPlexConfigTestSuite) TestRequiredFields() {
	var err error
	var pc = plexConfig{}
	pc.Enabled = true

	requiredFields := []string{
		"plexConfig.Token",
	}

	err = validate.Struct(pc)
	if assert.Error(suite.T(), err) {
		assert.IsType(suite.T(), validator.ValidationErrors{}, err)
		validationErrors := err.(validator.ValidationErrors)
		for i, field := range requiredFields {
			assert.Equal(suite.T(), "required_if", validationErrors[i].Tag())
			assert.Equal(suite.T(), field, validationErrors[i].Namespace())
		}
	}
}

func TestFolderRules(t *testing.T) {
	suite.Run(t, new(TestFolderRulesTestSuite))
}

func (suite *TestFolderRulesTestSuite) TestValidValues() {
	var err error
	var fr = FolderRules{}
	fr.DeleteWatchedAfterHours = 1
	fr.KeepShow = false
	fr.KeepSeasons = []int{1}

	err = validate.Struct(fr)
	assert.NoError(suite.T(), err)
}

func (suite *TestFolderRulesTestSuite) TestBothKeepShowAndKeepSeason() {
	var err error
	var fr = FolderRules{}
	fr.KeepShow = true
	fr.KeepSeasons = []int{}

	err = validate.Struct(fr)
	if assert.Error(suite.T(), err) {
		assert.IsType(suite.T(), validator.ValidationErrors{}, err)
		validationErrors := err.(validator.ValidationErrors)
		assert.Equal(suite.T(), "excluded_with_all", validationErrors[0].Tag())
		assert.Equal(suite.T(), "FolderRules.KeepShow", validationErrors[0].Namespace())
		assert.Equal(suite.T(), "excluded_with_all", validationErrors[1].Tag())
		assert.Equal(suite.T(), "FolderRules.KeepSeasons", validationErrors[1].Namespace())
	}
}

func (suite *TestFolderRulesTestSuite) TestKeepSeasonEmpty() {
	var err error
	var fr = FolderRules{}
	fr.KeepSeasons = []int{}

	err = validate.Struct(fr)
	if assert.Error(suite.T(), err) {
		assert.IsType(suite.T(), validator.ValidationErrors{}, err)
		validationErrors := err.(validator.ValidationErrors)
		assert.Equal(suite.T(), "gt", validationErrors[0].Tag())
		assert.Equal(suite.T(), "FolderRules.KeepSeasons", validationErrors[0].Namespace())
	}
}
