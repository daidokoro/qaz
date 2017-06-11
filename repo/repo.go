package repo

// All logic for Git clone and deploy commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"qaz/logger"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	billy "gopkg.in/src-d/go-billy.v2"
	"gopkg.in/src-d/go-billy.v2/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Repo used to manage git repo based deployments
type Repo struct {
	URL    string
	fs     *memfs.Memory
	Files  map[string]string
	Config string
	RSA    string
	User   string
	Secret string
}

// Log create Logger
var Log *logger.Logger

// NewRepo - returns pointer to a new repo struct
func NewRepo(url string) (*Repo, error) {
	r := &Repo{
		fs:    memfs.New(),
		Files: make(map[string]string),
		URL:   url,
	}

	if err := r.clone(); err != nil {
		return r, err
	}

	root, err := r.fs.ReadDir("/")
	if err != nil {
		return r, err
	}

	if err := r.readFiles(root, ""); err != nil {
		return r, err
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

	// set authentication
	if err := r.getAuth(opts); err != nil {
		return err
	}

	Log.Debug(fmt.Sprintln("calling [git clone] with params:", opts))

	// Clones the repository into the worktree (fs) and storer all the .git
	Log.Info(fmt.Sprintf("fetching git repo: [%s]\n--", filepath.Base(r.URL)))
	if _, err := git.Clone(store, r.fs, opts); err != nil {
		return err
	}

	fmt.Println("--")

	return nil
}

func (r *Repo) readFiles(root []billy.FileInfo, dirname string) error {
	Log.Debug(fmt.Sprintf("writing repo files to memory filesystem [%s]", dirname))
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
		r.Files[path] = buf.String()

	}
	return nil
}

func (r *Repo) getAuth(opts *git.CloneOptions) error {
	if strings.HasPrefix(r.URL, "git@") {
		Log.Debug("SSH Source URL detected, attempting to use SSH Keys")

		sshAuth, err := ssh.NewPublicKeysFromFile("git", r.RSA, "")
		if err != nil {
			return err
		}

		opts.Auth = sshAuth
		return nil
	}

	if r.User != "" {
		if r.Secret == "" {
			fmt.Printf("password:")
			p, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return err
			}
			fmt.Printf("\n")

			r.Secret = string(p)
		}
		opts.Auth = http.NewBasicAuth(r.User, r.Secret)
	}

	return nil
}
