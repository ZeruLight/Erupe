package tests

import (
	"os"
	"testing"

	"github.com/ktrysmt/go-bitbucket"
)

func getClient(t *testing.T) *bitbucket.Client {
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

func getOwner(t *testing.T) string {
	owner := os.Getenv("BITBUCKET_TEST_OWNER")

	if owner == "" {
		t.Error("BITBUCKET_TEST_OWNER is empty.")
	}

	return owner
}

func TestProjectCreate_isPrivateTrue(t *testing.T) {
	c := getClient(t)

	projectName := "go-bitbucket-test-project-create-is-private-true"
	projectKey := "GO_BB_TEST_PROJ_CR_IS_PRIV_TRUE"
	opt := &bitbucket.ProjectOptions{
		Owner:     getOwner(t),
		Name:      projectName,
		Key:       projectKey,
		IsPrivate: true,
	}
	project, err := c.Workspaces.CreateProject(opt)
	if err != nil {
		t.Error("The project could not be created.")
	}

	if project.Name != projectName {
		t.Error("The project `Name` attribute does not match the expected value.")
	}
	if project.Key != projectKey {
		t.Error("The project `Key` attribute does not match the expected value.")
	}
	if project.Is_private != true {
		t.Error("The project `Is_private` attribute does not match the expected value.")
	}

	// Delete the project, so we can keep a clean test environment
	_, err = c.Workspaces.DeleteProject(opt)
	if err != nil {
		t.Error(err)
	}
}

func TestProjectCreate_isPrivateFalse(t *testing.T) {
	c := getClient(t)

	projectName := "go-bitbucket-test-project-create-is-private-false"
	projectKey := "GO_BB_TEST_PROJ_CR_IS_PRIV_FALSE"
	opt := &bitbucket.ProjectOptions{
		Owner:     getOwner(t),
		Name:      projectName,
		Key:       projectKey,
		IsPrivate: false,
	}
	project, err := c.Workspaces.CreateProject(opt)
	if err != nil {
		t.Error("The project could not be created.")
	}

	if project.Name != projectName {
		t.Error("The project `Name` attribute does not match the expected value.")
	}
	if project.Key != projectKey {
		t.Error("The project `Key` attribute does not match the expected value.")
	}
	if project.Is_private != false {
		t.Error("The project `Is_private` attribute does not match the expected value.")
	}

	// Delete the project, so we can keep a clean test environment
	_, err = c.Workspaces.DeleteProject(opt)
	if err != nil {
		t.Error(err)
	}
}
