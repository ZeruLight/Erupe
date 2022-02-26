package tests

import (
	"os"
	"strconv"
	"testing"

	"github.com/ktrysmt/go-bitbucket"
)

var (
	user  = os.Getenv("BITBUCKET_TEST_USERNAME")
	pass  = os.Getenv("BITBUCKET_TEST_PASSWORD")
	owner = os.Getenv("BITBUCKET_TEST_OWNER")
	repo  = os.Getenv("BITBUCKET_TEST_REPOSLUG")
)

func setup(t *testing.T) *bitbucket.Client {

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
	return c
}

func TestBranchRestrictionsKindPush(t *testing.T) {

	c := setup(t)
	var pushRestrictionID int

	t.Run("create", func(t *testing.T) {
		opt := &bitbucket.BranchRestrictionsOptions{
			Owner:    owner,
			Pattern:  "develop",
			RepoSlug: repo,
			Kind:     "push",
			Users:    []string{user},
		}
		res, err := c.Repositories.BranchRestrictions.Create(opt)
		if err != nil {
			t.Error(err)
		}
		if res.Kind != "push" {
			t.Error("did not match branchrestriction kind")
		}
		pushRestrictionID = res.ID
	})

	t.Run("delete", func(t *testing.T) {
		opt := &bitbucket.BranchRestrictionsOptions{
			Owner:    owner,
			RepoSlug: repo,
			ID:       strconv.Itoa(pushRestrictionID),
		}
		_, err := c.Repositories.BranchRestrictions.Delete(opt)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestBranchRestrictionsKindRequirePassingBuilds(t *testing.T) {

	c := setup(t)
	var pushRestrictionID int

	t.Run("create", func(t *testing.T) {
		opt := &bitbucket.BranchRestrictionsOptions{
			Owner:    owner,
			Pattern:  "master",
			RepoSlug: repo,
			Kind:     "require_passing_builds_to_merge",
			Value:    2,
		}
		res, err := c.Repositories.BranchRestrictions.Create(opt)
		if err != nil {
			t.Error(err)
		}
		if res.Kind != "require_passing_builds_to_merge" {
			t.Error("did not match branchrestriction kind")
		}
		pushRestrictionID = res.ID
	})

	t.Run("delete", func(t *testing.T) {
		opt := &bitbucket.BranchRestrictionsOptions{
			Owner:    owner,
			RepoSlug: repo,
			ID:       strconv.Itoa(pushRestrictionID),
		}
		_, err := c.Repositories.BranchRestrictions.Delete(opt)
		if err != nil {
			t.Error(err)
		}
	})
}
