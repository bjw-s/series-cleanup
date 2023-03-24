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
	testExpectedValidationErrors(
		suite.T(),
		&traktConfig{
			Enabled:      true,
			ClientID:     "12345",
			ClientSecret: "12345",
			User:         "12345",
			CacheDir:     workingDir,
		},
		nil,
	)
}

func (suite *TestTraktConfigTestSuite) TestRequiredFields() {
	testExpectedValidationErrors(
		suite.T(),
		&traktConfig{
			Enabled: true,
		},
		[]*expectedValidationError{
			{tag: "required_if", namespace: "traktConfig.CacheDir"},
			{tag: "required_if", namespace: "traktConfig.ClientID"},
			{tag: "required_if", namespace: "traktConfig.ClientSecret"},
			{tag: "required_if", namespace: "traktConfig.User"},
		},
	)
}

func (suite *TestTraktConfigTestSuite) TestInvalidCacheFolder() {
	testExpectedValidationErrors(
		suite.T(),
		&traktConfig{
			Enabled:      true,
			ClientID:     "12345",
			ClientSecret: "12345",
			User:         "12345",
			CacheDir:     "/this_does_not_exist",
		},
		[]*expectedValidationError{
			{tag: "valid_folder", namespace: "traktConfig.CacheDir"},
		},
	)
}

func TestPlexConfig(t *testing.T) {
	suite.Run(t, new(TestPlexConfigTestSuite))
}

func (suite *TestPlexConfigTestSuite) TestDisabled() {
	testExpectedValidationErrors(
		suite.T(),
		&plexConfig{
			Enabled: false,
		},
		nil,
	)
}

func (suite *TestPlexConfigTestSuite) TestValidValues() {
	testExpectedValidationErrors(
		suite.T(),
		&plexConfig{
			Enabled: false,
			Token:   "12345",
		},
		nil,
	)
}

func (suite *TestPlexConfigTestSuite) TestRequiredFields() {
	testExpectedValidationErrors(
		suite.T(),
		&plexConfig{
			Enabled: true,
		},
		[]*expectedValidationError{
			{tag: "required_if", namespace: "plexConfig.Token"},
		},
	)
}

func TestFolderRules(t *testing.T) {
	suite.Run(t, new(TestFolderRulesTestSuite))
}

func (suite *TestFolderRulesTestSuite) TestValidValues() {
	testExpectedValidationErrors(
		suite.T(),
		&FolderRules{
			DeleteWatchedAfterHours: 1,
			KeepShow:                false,
			KeepSeasons:             []int{1},
		},
		nil,
	)
}

func (suite *TestFolderRulesTestSuite) TestBothKeepShowAndKeepSeason() {
	testExpectedValidationErrors(
		suite.T(),
		&FolderRules{
			KeepShow:    true,
			KeepSeasons: []int{},
		},
		[]*expectedValidationError{
			{tag: "excluded_with_all", namespace: "FolderRules.KeepShow"},
			{tag: "excluded_with_all", namespace: "FolderRules.KeepSeasons"},
		},
	)
}

func (suite *TestFolderRulesTestSuite) TestKeepSeasonEmpty() {
	testExpectedValidationErrors(
		suite.T(),
		&FolderRules{
			KeepSeasons: []int{},
		},
		[]*expectedValidationError{
			{tag: "gt", namespace: "FolderRules.KeepSeasons"},
		},
	)
}

func testExpectedValidationErrors(t *testing.T, obj interface{}, ev []*expectedValidationError) {
	var err = validate.Struct(obj)
	if ev == nil {
		assert.NoError(t, err)
	} else {
		if assert.Error(t, err) {
			assert.IsType(t, validator.ValidationErrors{}, err)
			validationErrors := err.(validator.ValidationErrors)
			for i, validationError := range ev {
				assert.Equal(t, validationError.tag, validationErrors[i].Tag())
				assert.Equal(t, validationError.namespace, validationErrors[i].Namespace())
			}
		}
	}
}

type expectedValidationError struct {
	tag       string
	namespace string
}
