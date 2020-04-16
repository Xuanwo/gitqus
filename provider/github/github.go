package github

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

type Client struct {
	c *github.Client

	owner  string
	repo   string
	branch string

	authorName  string
	authorEmail string
}

func New(owner, repo, branch, authorName, authorEmail string) *Client {
	ctx := context.Background()
	oc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITQUS_GITHUB_ACCESS_TOKEN")}))

	return &Client{
		c: github.NewClient(oc),

		owner:       owner,
		repo:        repo,
		branch:      branch,
		authorName:  authorName,
		authorEmail: authorEmail,
	}
}

func (c *Client) GetFile(ctx context.Context, path string) (_ []byte, err error) {
	file, _, _, err := c.c.Repositories.GetContents(ctx, c.owner, c.repo, path, nil)
	if err != nil {
		var e *github.ErrorResponse
		if ok := errors.As(err, &e); ok {
			if e.Response.StatusCode == http.StatusNotFound {
				return nil, nil
			}
		}
		return nil, err
	}

	content, err := file.GetContent()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (c *Client) CreateOrUpdateFile(ctx context.Context, path, message, content string) (branch string, err error) {
	var baseRef *github.Reference
	if baseRef, _, err = c.c.Git.GetRef(ctx, c.owner, c.repo, "refs/heads/"+c.branch); err != nil {
		return "", err
	}

	refName := "refs/heads/" + strings.ReplaceAll(path, "/", "_")
	var ref *github.Reference
	ref, _, _ = c.c.Git.GetRef(ctx, c.owner, c.repo, refName)
	if ref == nil {
		ref, _, err = c.c.Git.CreateRef(ctx, c.owner, c.repo, &github.Reference{Ref: github.String(refName), Object: &github.GitObject{SHA: baseRef.Object.SHA}})
		if err != nil {
			return "", err
		}
	}

	// Create a tree with what to commit.
	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(path), Type: github.String("blob"), Content: github.String(content), Mode: github.String("100644")})
	tree, _, err := c.c.Git.CreateTree(ctx, c.owner, c.repo, *ref.Object.SHA, entries)
	if err != nil {
		return "", err
	}

	parent, _, err := c.c.Repositories.GetCommit(ctx, c.owner, c.repo, *ref.Object.SHA)
	if err != nil {
		return "", err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &c.authorName, Email: &c.authorEmail}
	commit := &github.Commit{Author: author, Message: github.String(message), Tree: tree, Parents: []*github.Commit{parent.Commit}}
	newCommit, _, err := c.c.Git.CreateCommit(ctx, c.owner, c.repo, commit)
	if err != nil {
		log.Printf("failed to create commit: %v", err)
		return "", err
	}
	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = c.c.Git.UpdateRef(ctx, c.owner, c.repo, ref, false)
	if err != nil {
		return "", err
	}
	return ref.GetRef(), nil
}

func (c *Client) CreatePR(ctx context.Context, title, head, base string) (number int, err error) {
	pr, _, err := c.c.PullRequests.Create(ctx, c.owner, c.repo, &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(head),
		Base:                github.String(base),
		MaintainerCanModify: github.Bool(false),
	})
	if err != nil {
		return
	}
	return *pr.Number, nil
}

func (c *Client) MergePR(ctx context.Context, number int) (err error) {
	_, _, err = c.c.PullRequests.Merge(ctx, c.owner, c.repo, number, "", &github.PullRequestOptions{
		MergeMethod: "squash",
	})
	if err != nil {
		return
	}
	return
}
