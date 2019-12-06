package bugtracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const baseURL = "https://%s.atlassian.net/rest/api/3/issue/%s"

// known jira errors
const (
	jiraNotExist = `Issue does not exist or you do not have permission to see it.`
)

// Errors
var (
	ErrTicketNotExist = errors.New("issue does not exist")
)

type Ticketer interface {
	Name(issueID string) (name string, err error)
}

type Jira struct {
	Site  string `json:"site"`
	Email string `json:"email"`
	Token string `json:"token"`

	nameCache map[string]*string
}

type Ticket struct {
	Fields    Fields   `json:"fields"`
	ErrorMsgs []string `json:"errorMessages"`
}

type Fields struct {
	Summary string `json:"summary"`
}

// NewJira creates a new Jira instance.
func NewJira(site, email, token string) *Jira {
	return &Jira{
		Site:  site,
		Email: email,
		Token: token,

		nameCache: make(map[string]*string),
	}
}

// JiraFromFile creates Jira instance with auth details from the file.
func JiraFromFile(filename string) (*Jira, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	j := NewJira("", "", "")
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}
	if j.Site == "" || j.Email == "" || j.Token == "" {
		return nil, errors.New("either site, email or token are missing from the config")
	}
	return j, nil
}

// Name returns the name of the ticket
func (j *Jira) Name(ticket string) (string, error) {
	name, ok := j.nameCache[ticket]
	if ok {
		if name == nil {
			return "", ErrTicketNotExist
		}
		return *name, nil
	}
	uri := fmt.Sprintf(baseURL, j.Site, ticket)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(j.Email, j.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	t := Ticket{}
	if err := json.Unmarshal(data, &t); err != nil {
		return "", err
	}
	if t.ErrorMsgs != nil {
		msg := t.ErrorMsgs[0]
		if strings.Compare(msg, jiraNotExist) == 0 {
			j.nameCache[ticket] = nil
			return "", ErrTicketNotExist
		}
		return "", errors.New(msg)
	}
	j.nameCache[ticket] = &t.Fields.Summary
	return t.Fields.Summary, nil
}
