package tests

import (
	"os"
	"fmt"
	"testing"
	"time"

	_ "github.com/k0kubun/pp"
	"github.com/ktrysmt/go-bitbucket"
)

func TestEndToEndDeploymentVariables(t *testing.T) {

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

	environmentOpt := &bitbucket.RepositoryEnvironmentsOptions{
		Owner:    owner,
		RepoSlug: repo,
	}

	environments, err := c.Repositories.Repository.ListEnvironments(environmentOpt)
	if err != nil {
		t.Error(err)
	}

	if environments == nil {
		t.Error("list didn't return any environments")
	}

	environment, err := findEnvironmentByName("Test", environments)
	if err != nil {
		t.Error(err)
	}

	opt := &bitbucket.RepositoryDeploymentVariableOptions{
		Owner:       owner,
		RepoSlug:    repo,
		Environment: environment,
		Key:         "foo",
		Value:       "value",
	}

	variable, err := c.Repositories.Repository.AddDeploymentVariable(opt)
	if err != nil {
		t.Error(err)
	}

	opt.Key = "bar"
	opt.Value = "other value"
	opt.Uuid = variable.Uuid

	updatedVariable, err := c.Repositories.Repository.UpdateDeploymentVariable(opt)
	if err != nil {
		t.Error(err)
	}

	listOpt := &bitbucket.RepositoryDeploymentVariablesOptions{
		Owner:       owner,
		RepoSlug:    repo,
		Environment: environment,
		MaxDepth:    10,
		PageNum:     1,
		Pagelen:     10,
	}

	err = waitForVariables(updatedVariable.Uuid, "bar", "other value", c, listOpt)
	if err != nil {
		t.Error(err)
	}

	deleteOpt := &bitbucket.RepositoryDeploymentVariableDeleteOptions{
		Owner:       owner,
		RepoSlug:    repo,
		Environment: environment,
		Uuid:        variable.Uuid,
	}

	_, err = c.Repositories.Repository.DeleteDeploymentVariable(deleteOpt)
	if err != nil {
		t.Error(err)
	}

	err = waitForDeletion(updatedVariable.Uuid, c, listOpt)
	if err == nil {
		t.Error("updated variable was not deleted")
	}
}

func findEnvironmentByName(name string, e *bitbucket.Environments) (*bitbucket.Environment, error) {
	for _, environment := range e.Environments {
		if environment.Name == name {
			return &environment, nil
		}
	}

	return nil, fmt.Errorf("no environment named %s", name)
}

func waitForDeletion(uuid string, c *bitbucket.Client, opt *bitbucket.RepositoryDeploymentVariablesOptions) error {
	for i := 0; i < 3; i++ {
		deploymentVariables, err := c.Repositories.Repository.ListDeploymentVariables(opt)
		if err != nil {
			return err
		}

		for _, deploymentVariable := range deploymentVariables.Variables {
			if deploymentVariable.Uuid == uuid {
				time.Sleep(3 * time.Second)

				break
			}
		}
	}

	return fmt.Errorf("update variable not found in list")
}

func waitForVariables(uuid string, key string, value string, c *bitbucket.Client, opt *bitbucket.RepositoryDeploymentVariablesOptions) error {
	for i := 0; i < 3; i++ {
		deploymentVariables, err := c.Repositories.Repository.ListDeploymentVariables(opt)
		if err != nil {
			return err
		}

		for _, deploymentVariable := range deploymentVariables.Variables {
			if deploymentVariable.Uuid == uuid && deploymentVariable.Key == key && deploymentVariable.Value == value {
				return nil
			}
		}

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("update variable not found in list")
}
