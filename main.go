package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	term "github.com/nsf/termbox-go"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type Repos struct {
	Name string `json:"name"`
}

type MenuState struct {
	selectedIndex int
	repos         []Repos
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
	term.Sync() // cosmestic purpose
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
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	token := "507fe7c80e1024e93350934a2bdf775056bc7801"
	gitHost := "https://git.campmon.com/"
	gitAPI := gitHost + "api/v3/"

	req, err := http.NewRequest("GET", gitAPI+"users/kenleyb/repos", nil)
	req.Header.Add("Authorization", `Bearer `+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("WDF")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("WDF")
		return
	}
	var arr []Repos
	_ = json.Unmarshal(body, &arr)
	// log.Printf("Unmarshaled: %v", arr)

	state := &MenuState{
		repos:         arr,
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
	runCloner(state)
	fmt.Println("Done.")
}

func runCloner(state *MenuState) {
	repoName := state.repos[state.selectedIndex].Name
	fmt.Println("\nSelected Repo -", repoName)
	fmt.Println("Cloning: ", repoName)
	auth, err := getGitAuth()
	if err != nil {
		fmt.Println("Authentication Failed. Please add you ssh public key to $HOME/.ssh/id_rsa.")
	}

	os.RemoveAll("./tmp")
	srcDir := "tmp/carbon/"
	repoDir := srcDir + repoName
	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:      "git@git.campmon.com:kenleyb/" + repoName + ".git",
		Auth:     auth,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Println("WDF", err)
		return
	}

	// Parse generator config
	raw, err := ioutil.ReadFile(repoDir + "/carbon.json")
	if err != nil {
		fmt.Println("carbon.json doesn't exist, couldn't template this project.")
		os.Exit(1)
	}

	var prompts []PromptConfig
	json.Unmarshal(raw, &prompts)

	// Prompt for template input
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nProject name: ")
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

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

	// Copy to final dest
	//  destPath := "./" + projectName

	e := filepath.Walk(repoDir, func(path string, f os.FileInfo, err error) error {
		// fmt.Println(strings.TrimPrefix(path, repoDir))
		destPath := projectName + strings.TrimPrefix(path, repoDir)
		err = TemplateFile(path, destPath, templateMap)
		return err
	})

	if e != nil {
		panic(e)
	}

	// Apply templating
	a, err := template.New("test").Parse("\n{{.moduleName}} are made out of {{.description}}\n")
	if err != nil {
		panic(err)
	}
	err = a.Execute(os.Stdout, templateMap)
	if err != nil {
		panic(err)
	}

	// Copy to final dest
}

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

// TemplateFile copies a file from src to dst and applies text templating.
func TemplateFile(src, dst string, templateMap map[string]string) (err error) {
	_, err = os.Stat(src)
	if err != nil {
		return
	}

	// srcFile, err := os.Open(src)
	// if err != nil {
	// 	return
	// }
	// defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	// _, err = io.Copy(dstFile, srcFile)
	// if err != nil {
	// 	return
	// }

	// Template work
	w := bufio.NewWriter(dstFile)
	tmpl, err := template.ParseFiles(src)
	if err != nil {
		return
	}
	err = tmpl.ExecuteTemplate(w, src, templateMap)
	if err != nil {
		return
	}

	// err = dstFile.Sync()
	// if err != nil {
	// 	return
	// }

	// err = os.Chmod(dst, si.Mode())
	// if err != nil {
	// 	return
	// }

	return
}
