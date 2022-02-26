package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/ktrysmt/go-bitbucket"
)

func TestList(t *testing.T) {

	user := os.Getenv("BITBUCKET_TEST_USERNAME")
	pass := os.Getenv("BITBUCKET_TEST_PASSWORD")
	owner := os.Getenv("BITBUCKET_TEST_OWNER")

	c := bitbucket.NewBasicAuth(user, pass)

	opt := &bitbucket.RepositoriesOptions{
		Owner: owner,
	}

	repos, err := c.Repositories.ListForAccount(opt)

	if err != nil {
		t.Error(err)
	}

	for _, repo := range repos.Items {

		tagOpt := &bitbucket.RepositoryTagOptions{
			Owner:    owner,
			RepoSlug: repo.Slug,
		}
		res, err := c.Repositories.Repository.ListTags(tagOpt)

		if err != nil {
			t.Error(err)
		}

		if len(res.Tags) != 0 {
			for _, tag := range res.Tags {
				fmt.Println(tag.Name)
			}
		}
	}
}
