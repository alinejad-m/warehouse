package gitrepo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo runs git commands in WorkDir.
type Repo struct {
	WorkDir string
	Branch  string
	// Remote is the remote name for pull/push (default origin).
	Remote string
	Author string
	Email  string
}

func (r *Repo) remoteName() string {
	if r.Remote != "" {
		return r.Remote
	}
	return "origin"
}

func (r *Repo) git(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.WorkDir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+r.Author,
		"GIT_COMMITTER_NAME="+r.Author,
		"GIT_AUTHOR_EMAIL="+r.Email,
		"GIT_COMMITTER_EMAIL="+r.Email,
	)
	return cmd
}

// Exists reports whether WorkDir looks like a git repository.
func (r *Repo) Exists() bool {
	st, err := os.Stat(filepath.Join(r.WorkDir, ".git"))
	return err == nil && st.IsDir()
}

// GetRemoteURL returns the fetch URL for the named remote (e.g. origin).
func GetRemoteURL(workDir, remote string) (string, error) {
	if remote == "" {
		remote = "origin"
	}
	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url %s: %w", remote, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Clone runs git clone into parent of WorkDir (WorkDir must not exist) or empty dir.
func Clone(remoteURL, workDir, branch string) error {
	parent := filepath.Dir(workDir)
	base := filepath.Base(workDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	args := []string{"clone", "--branch", branch, "--single-branch", remoteURL, base}
	cmd := exec.Command("git", args...)
	cmd.Dir = parent
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// InitOrClone clones remoteURL into workDir when .git is missing.
func InitOrClone(remoteURL, workDir, branch, remote, author, email string) (*Repo, error) {
	repo := &Repo{WorkDir: workDir, Branch: branch, Remote: remote, Author: author, Email: email}
	if repo.Exists() {
		return repo, nil
	}
	if err := os.MkdirAll(filepath.Dir(workDir), 0o755); err != nil {
		return nil, err
	}
	if err := Clone(remoteURL, workDir, branch); err != nil {
		return nil, err
	}
	return repo, nil
}

// Pull runs git pull --ff-only for the configured branch.
func (r *Repo) Pull() error {
	out, err := r.git("pull", r.remoteName(), r.Branch, "--ff-only").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Add runs git add on paths (repo-relative or absolute under workdir).
func (r *Repo) Add(paths ...string) error {
	args := append([]string{"add"}, paths...)
	out, err := r.git(args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Remove runs git rm --cached or normal rm tracked files.
func (r *Repo) Remove(force bool, paths ...string) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, paths...)
	out, err := r.git(args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git rm: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Commit creates a commit if there is anything staged.
func (r *Repo) Commit(message string) error {
	err := r.git("diff", "--cached", "--quiet").Run()
	if err == nil {
		return nil
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 1 {
		return fmt.Errorf("git diff --cached: %w", err)
	}
	out, err := r.git("commit", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Push runs git push <remote> branch.
func (r *Repo) Push() error {
	out, err := r.git("push", r.remoteName(), r.Branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// StatusPorcelain returns true if work tree or index has changes.
func (r *Repo) StatusPorcelain() (bool, error) {
	out, err := r.git("status", "--porcelain").CombinedOutput()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
