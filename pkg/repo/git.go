package repo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// A GitCollection represents a git's collection type of repos (e.g users or orgs)
type GitCollection string

// The defined type for the available repo collection
const (
	User         GitCollection = "users"
	Organization               = "orgs"
)

type Repos struct {
	Name string `json:"name"`
}

type ReposLister interface {
	GetRepos() ([]Repos, error)
	CloneRepo(repoName string, dest string) error
}

type gitConfig struct {
	token          string
	apiURL         *url.URL
	collection     GitCollection
	collectionName string
}

type gitCollection struct {
	config gitConfig
}

// NewCollection creates a new git repos collection
func NewCollection(apiURL *url.URL, token, collectionName string, collectionType GitCollection) ReposLister {
	return &gitCollection{config: gitConfig{token: token, apiURL: apiURL, collectionName: collectionName, collection: collectionType}}
}

func (g *gitCollection) GetRepos() (repos []Repos, err error) {
	url := g.config.apiURL.String() + "/" + string(g.config.collection) + "/" + g.config.collectionName + "/repos"
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", `Bearer `+g.config.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var arr []Repos
	_ = json.Unmarshal(body, &arr)
	return arr, nil
}

// getGitAuth grabs git authentication key from home dir, TODO: windows
func getGitAuth() (*gitssh.PublicKeys, error) {
	s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	sshKey, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return nil, err
	}
	auth := &gitssh.PublicKeys{User: "git", Signer: signer}
	return auth, nil
}

func (g *gitCollection) CloneRepo(repoName string, dest string) (err error) {
	auth, err := getGitAuth()
	if err != nil {
		err = fmt.Errorf("Authentication Failed. Please add your ssh public key to $HOME/.ssh/id_rsa. %s", err)
	}

	_, err = git.PlainClone(dest, false, &git.CloneOptions{
		URL:      "git@git.campmon.com:" + g.config.collectionName + "/" + repoName + ".git",
		Auth:     auth,
		Progress: os.Stdout,
	})
	return
}
