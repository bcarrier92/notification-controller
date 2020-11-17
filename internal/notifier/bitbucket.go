/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package notifier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fluxcd/pkg/recorder"
	"github.com/ktrysmt/go-bitbucket"
)

type Bitbucket struct {
	Owner  string
	Repo   string
	Client *bitbucket.Client
}

func NewBitbucket(addr string, token string) (*Bitbucket, error) {
	if len(token) == 0 {
		return nil, errors.New("Bitbucket token cannot be empty")
	}

	_, id, err := parseGitAddress(addr)
	if err != nil {
		return nil, err
	}

	comp := strings.Split(token, ":")
	if len(comp) != 2 {
		return nil, errors.New("Invalid token format, expected to be <user>:<password>")
	}
	username := comp[0]
	password := comp[1]

	comp = strings.Split(id, "/")
	if len(comp) != 2 {
		return nil, fmt.Errorf("Invalid repository id '%s'", id)
	}
	owner := comp[0]
	repo := comp[1]

	return &Bitbucket{
		Owner:  owner,
		Repo:   repo,
		Client: bitbucket.NewBasicAuth(username, password),
	}, nil
}

// Post Bitbucket commit status
func (b Bitbucket) Post(event recorder.Event) error {
	// Skip progressing events
	if event.Reason == "Progressing" {
		return nil
	}

	revString, ok := event.Metadata["revision"]
	if !ok {
		return errors.New("Missing revision metadata")
	}
	rev, err := parseRevision(revString)
	if err != nil {
		return err
	}
	state, err := toBitbucketState(event.Severity)
	if err != nil {
		return err
	}

	name, desc := formatNameAndDescription(event)

	cmo := &bitbucket.CommitsOptions{
		Owner:    b.Owner,
		RepoSlug: b.Repo,
		Revision: rev,
	}
	cso := &bitbucket.CommitStatusOptions{
		State:       state,
		Key:         name,
		Name:        name,
		Description: desc,
		Url:         "https://bitbucket.org",
	}
	_, err = b.Client.Repositories.Commits.CreateCommitStatus(cmo, cso)
	if err != nil {
		return err
	}

	return nil
}

func toBitbucketState(severity string) (string, error) {
	switch severity {
	case recorder.EventSeverityInfo:
		return "SUCCESSFUL", nil
	case recorder.EventSeverityError:
		return "FAILED", nil
	default:
		return "", errors.New("Can't convert to GitHub state")
	}
}
