package commands

// All logic for Git clone and deploy commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	billy "gopkg.in/src-d/go-billy.v2"
	"gopkg.in/src-d/go-billy.v2/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Repo used to manage git repo based deployments
type Repo struct {
	URL    string
	fs     *memfs.Memory
	files  map[string]string
	config string
}

var gitrepo Repo

// NewRepo - returns pointer to a new repo struct
func NewRepo(url string) (*Repo, error) {
	r := &Repo{
		fs:    memfs.New(),
		files: make(map[string]string),
		URL:   url,
	}

	if err := r.clone(); err != nil {
		return r, err
	}

	root, err := r.fs.ReadDir("/")
	if err != nil {
		handleError(err)
		return r, nil
	}

	if err := r.readFiles(root, ""); err != nil {
		handleError(err)
		return r, nil
	}

	return r, nil
}

func (r *Repo) clone() error {
	// memory store for git objects
	store := memory.NewStorage()

	// clone options
	opts := &git.CloneOptions{
		URL:      r.URL,
		Progress: os.Stdout,
	}

	if run.gituser != "" {
		if run.gitpass == "" {
			fmt.Printf("password:")
			p, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return err
			}
			fmt.Printf("\n")

			run.gitpass = string(p)
		}
		opts.Auth = http.NewBasicAuth(run.gituser, run.gitpass)
	}

	Log(fmt.Sprintln("calling [git clone] with params:", opts), level.debug)

	// Clones the repository into the worktree (fs) and storer all the .git
	Log(fmt.Sprintf("fetching git repo: [%s]\n--", filepath.Base(r.URL)), level.info)
	if _, err := git.Clone(store, r.fs, opts); err != nil {
		return err
	}

	fmt.Println("--")

	return nil
}

func (r *Repo) readFiles(root []billy.FileInfo, dirname string) error {
	Log(fmt.Sprintf("writing repo files to memory filesystem [%s]", dirname), level.debug)
	for _, i := range root {
		if i.IsDir() {
			dir, _ := r.fs.ReadDir(i.Name())
			r.readFiles(dir, i.Name())
			continue
		}

		path := filepath.Join(dirname, i.Name())

		out, err := r.fs.Open(path)
		if err != nil {
			return err
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(out)

		// update file map
		r.files[path] = buf.String()

	}
	return nil
}
