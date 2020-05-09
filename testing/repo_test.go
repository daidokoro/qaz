package testing

import (
	"fmt"
	"testing"

	"github.com/daidokoro/qaz/repo"
	"github.com/stretchr/testify/assert"
)

const repoURI = "https://github.com/cfn-deployable/simplevpc"

// TestRepo - test repo package
func TestRepo(t *testing.T) {
	// t.Parallel()
	fmt.Println("running")
	repo, err := repo.New(repoURI, "", "nada")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(repo.Files))
	for k := range repo.Files {
		t.Log("repo file:", k)
	}
	// fmrepo.Files
}
