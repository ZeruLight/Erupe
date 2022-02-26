package tests

import (
	"os"
	"testing"

	"github.com/ktrysmt/go-bitbucket"
)

func getBitbucketClient(t *testing.T) *bitbucket.Client {
	user := os.Getenv("BITBUCKET_TEST_USERNAME")
	pass := os.Getenv("BITBUCKET_TEST_PASSWORD")

	if user == "" {
		t.Error("BITBUCKET_TEST_USERNAME is empty.")
	}
	if pass == "" {
		t.Error("BITBUCKET_TEST_PASSWORD is empty.")
	}

	return bitbucket.NewBasicAuth(user, pass)
}

func getWorkspace(t *testing.T) string {
	owner := os.Getenv("BITBUCKET_TEST_OWNER")
	if owner == "" {
		t.Error("BITBUCKET_TEST_OWNER is empty.")
	}

	return owner
}

func getRepositoryName(t *testing.T) string {
	repo := os.Getenv("BITBUCKET_TEST_REPOSLUG")
	if repo == "" {
		t.Error("BITBUCKET_TEST_REPOSLUG is empty.")
	}

	return repo
}

func TestListWorkspaces(t *testing.T) {
	c := getBitbucketClient(t)

	res, err := c.Workspaces.List()
	if err != nil {
		t.Error("The workspaces could not be listed.")
	}

	if res == nil || res.Size == 0 {
		t.Error("Should have at least one workspace")
	}
}

func TestGetWorkspace(t *testing.T) {
	c := getBitbucketClient(t)
	workspaceName := getWorkspace(t)

	res, err := c.Workspaces.Get(workspaceName)
	if err != nil {
		t.Error("Could not get the workspace.")
	}

	if res == nil || res.Slug != workspaceName {
		t.Error("The workspace was not returned")
	}
}

func TestGetWorkspaceRepository(t *testing.T) {
	c := getBitbucketClient(t)
	workspaceName := getWorkspace(t)
	repositoryName := getRepositoryName(t)

	opt := &bitbucket.RepositoryOptions{
		Owner:    workspaceName,
		RepoSlug: repositoryName,
	}

	res, err := c.Workspaces.Repositories.Repository.Get(opt)
	if err != nil {
		t.Error("The repository is not found.")
	}

	if res.Full_name != workspaceName+"/"+repositoryName {
		t.Error("Cannot catch repos full name.")
	}
}

func TestGetWorkspacePermissionForUser(t *testing.T) {
	c := getBitbucketClient(t)
	workspaceName := getWorkspace(t)

	user, err := c.User.Profile()
	if err != nil {
		t.Error(err)
	}

	res, err := c.Workspaces.Permissions.GetUserPermissionsByUuid(workspaceName, user.Uuid)
	if err != nil {
		t.Error("Could not get the workspace.")
	}

	if res == nil || res.Type == "" {
		t.Error("The workspace was not returned")
	}
}

func TestGetWorkspaceProjects(t *testing.T) {
	c := getBitbucketClient(t)
	workspaceName := getWorkspace(t)

	res, err := c.Workspaces.Projects(workspaceName)
	if err != nil {
		t.Error("could not get workspace projects")
	}

	if res == nil || len(res.Items) == 0 {
		t.Error("no workspace projects were returned")
	}
}
