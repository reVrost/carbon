package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"git.campmon.com/kenleyb/carbon/pkg/tmpl"

	repo "git.campmon.com/kenleyb/carbon/pkg/repo"
	"github.com/fatih/color"
	term "github.com/nsf/termbox-go"
)

type MenuState struct {
	selectedIndex int
	repos         []repo.Repos
}

type PromptConfig struct {
	Name         string `json:"name"`
	Message      string `json:"message"`
	DefaultValue string `json:"default_value,omitempty"`
	PromptType   string `json:"prompt_type,omitempty"`
}

func printTitle() {
	d := color.New(color.FgCyan, color.Bold)
	d.Print(`   ___           _                 
  / __\__ _ _ __| |__   ___  _ __  
 / /  / _' | '__| '_ \ / _ \| '_ \ 
/ /__| (_| | |  | |_) | (_) | | | |
\____/\__,_|_|  |_.__/ \___/|_| |_|

`)
	fmt.Println("  Project generator for the lazy dudes.")
}

func drawMenu(state *MenuState, delta int) {
	term.Sync()
	printTitle()
	d := color.New(color.FgYellow, color.Bold)
	u := color.New(color.FgWhite, color.Underline)
	u.Println("\nSelect from the available generators:")

	state.selectedIndex = (state.selectedIndex + delta) % len(state.repos)

	for i, x := range state.repos {
		if i == state.selectedIndex {
			d.Println("ο " + x.Name)
		} else {
			fmt.Println("ο " + x.Name)
		}
	}
}

func main() {
	apiURL, _ := url.Parse("https://git.campmon.com/api/v3")
	collection := repo.NewCollection(apiURL, "507fe7c80e1024e93350934a2bdf775056bc7801", "kenleyb", repo.User)
	fmt.Println("mama")
	repos, err := collection.GetRepos()
	if err != nil {
		fmt.Println(err)
	}
	if len(repos) == 0 {
		fmt.Println("No templates available")
		os.Exit(0)
	}

	state := &MenuState{
		repos:         repos,
		selectedIndex: 0,
	}

	term.Init()
	drawMenu(state, 0)
	isExitedByUser := false
eventLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				fallthrough
			case term.KeyCtrlC:
				fallthrough
			case term.KeyCtrlZ:
				isExitedByUser = true
				break eventLoop
			case term.KeyArrowUp:
				drawMenu(state, -1)
			case term.KeyArrowDown:
				drawMenu(state, 1)
			case term.KeyEnter:
				break eventLoop
			default:
				// we only want to read a single character or one key pressed event
				// fmt.Println("ASCII : ", ev.Ch)

			}
		case term.EventError:
			panic(ev.Err)
		}
	}
	term.Close()

	if isExitedByUser {
		fmt.Println("Exited by user.")
		os.Exit(0)
	}
	printTitle()
	runCloner(state, collection)
	fmt.Println("Done.")
}

func parsePromptConfig(repoDir string) ([]PromptConfig, error) {
	// Parse generator config (Config Parser)
	raw, err := ioutil.ReadFile(repoDir + "/carbon.json")
	if err != nil {
		return nil, err
	}

	var prompts []PromptConfig
	json.Unmarshal(raw, &prompts)
	return prompts, nil
}

func runCloner(state *MenuState, collection repo.ReposLister) {
	// Grabbing Repo (Repo Fetcher/Template Fetcher)
	repoName := state.repos[state.selectedIndex].Name
	fmt.Println("\nSelected Repo -", repoName)
	fmt.Println("Cloning: ", repoName)

	templatesDir := "templates/"
	os.RemoveAll(templatesDir)
	repoDir := templatesDir + repoName

	err := collection.CloneRepo(repoName, repoDir)
	if err != nil {
		fmt.Println("Couldn't clone repo "+repoName, err)
		return
	}
	os.RemoveAll(repoDir + "/.git")

	// Parse generator config (Config Parser)
	prompts, err := parsePromptConfig(repoDir)
	if err != nil {
		fmt.Println("carbon.json couldn't be read, can't template this project.")
		os.Exit(1)
	}

	// Prompt for template input
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nProject name: ")
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	if projectName == "" {
		fmt.Println("Project name cannot be empty")
		os.Exit(1)
	}

	if exists(projectName) {
		fmt.Println("Dir " + projectName + " already exists")
		os.Exit(1)
	}

	templateMap := make(map[string]string)
	for _, x := range prompts {

		prompt := x.Message + " "
		if x.DefaultValue != "" {
			prompt += "(" + x.DefaultValue + ") "
		}
		fmt.Printf(prompt)

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text != "" {
			templateMap[x.Name] = strings.TrimSpace(text)
		} else {
			templateMap[x.Name] = x.DefaultValue
		}
	}

	err = tmpl.ApplyDir(repoDir, projectName, templateMap)

	if err != nil {
		fmt.Println(err)
	}
}

// exists returns true if file/dir exists
func exists(filePath string) (exists bool) {
	exists = true

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	return
}
