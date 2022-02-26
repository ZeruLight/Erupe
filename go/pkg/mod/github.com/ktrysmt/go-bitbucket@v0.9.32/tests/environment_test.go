package tests

import (
	"os"
	"testing"

	_ "github.com/k0kubun/pp"
	"github.com/ktrysmt/go-bitbucket"
)

func TestListEnvironments(t *testing.T) {

	user := os.Getenv("BITBUCKET_TEST_USERNAME")
	pass := os.Getenv("BITBUCKET_TEST_PASSWORD")
	owner := os.Getenv("BITBUCKET_TEST_OWNER")
	repo := os.Getenv("BITBUCKET_TEST_REPOSLUG")

	if user == "" {
		t.Error("BITBUCKET_TEST_USERNAME is empty.")
	}
	if pass == "" {
		t.Error("BITBUCKET_TEST_PASSWORD is empty.")
	}
	if owner == "" {
		t.Error("BITBUCKET_TEST_OWNER is empty.")
	}
	if repo == "" {
		t.Error("BITBUCKET_TEST_REPOSLUG is empty.")
	}

	c := bitbucket.NewBasicAuth(user, pass)

	opt := &bitbucket.RepositoryEnvironmentsOptions{
		Owner:    owner,
		RepoSlug: repo,
	}

	res, err := c.Repositories.Repository.ListEnvironments(opt)
	if err != nil {
		t.Error(err)
	}

	if res == nil {
		t.Error("list didn't return any environments")
	}
}

func TestEndToEndEnvironments(t *testing.T) {

	user := os.Getenv("BITBUCKET_TEST_USERNAME")
	pass := os.Getenv("BITBUCKET_TEST_PASSWORD")
	owner := os.Getenv("BITBUCKET_TEST_OWNER")
	repo := os.Getenv("BITBUCKET_TEST_REPOSLUG")

	if user == "" {
		t.Error("BITBUCKET_TEST_USERNAME is empty.")
	}
	if pass == "" {
		t.Error("BITBUCKET_TEST_PASSWORD is empty.")
	}
	if owner == "" {
		t.Error("BITBUCKET_TEST_OWNER is empty.")
	}
	if repo == "" {
		t.Error("BITBUCKET_TEST_REPOSLUG is empty.")
	}

	c := bitbucket.NewBasicAuth(user, pass)

	opt := &bitbucket.RepositoryEnvironmentOptions{
		Owner:           owner,
		RepoSlug:        repo,
		Name:            "foo",
		EnvironmentType: bitbucket.Test,
	}

	environment, err := c.Repositories.Repository.AddEnvironment(opt)
	if err != nil {
		t.Error(err)
	}

	if environment.Uuid == "" {
		t.Error("new environment does not have a UUID")
	}

	opt.Name = ""
	opt.Uuid = environment.Uuid

	foundEnvironment, err := c.Repositories.Repository.GetEnvironment(opt)
	if err != nil {
		t.Error(err)
	}

	if foundEnvironment.Name != "foo" {
		t.Errorf("got environment with wrong name: %s", foundEnvironment.Name)
	}

	deleteOpt := &bitbucket.RepositoryEnvironmentDeleteOptions{
		Owner:    owner,
		RepoSlug: repo,
		Uuid:     environment.Uuid,
	}

	// On success the delete API doesn't return any content (HTTP status 204)
	_, err = c.Repositories.Repository.DeleteEnvironment(deleteOpt)
	if err != nil {
		t.Error(err)
	}
}
