package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/fatih/color"
	term "github.com/nsf/termbox-go"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type Repos struct {
	Name string `json:name`
}

type MenuState struct {
	selectedIndex int
	repos         []Repos
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
	term.Init()
	defer term.Close()
	printTitle()
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

	drawMenu(state, 0)
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				fmt.Println("Exited.")
				os.Exit(0)
			case term.KeyCtrlC:
				fmt.Println("Exited.")
				os.Exit(0)
			case term.KeyCtrlZ:
				fmt.Println("Exited.")
				os.Exit(0)
			case term.KeyArrowUp:
				drawMenu(state, -1)
			case term.KeyArrowDown:
				drawMenu(state, 1)
			case term.KeyEnter:
				fmt.Println("Enter")
				os.Exit(0)
				// drawMenu(state, term.KeyEnter)
			default:
				// we only want to read a single character or one key pressed event
				fmt.Println("ASCII : ", ev.Ch)

			}
		case term.EventError:
			panic(ev.Err)
		}
	}

	log.Printf("Cloning: %v", arr[0].Name)

	// auth, err := getGitAuth()
	// if err != nil {
	// 	fmt.Println("Authentication Failed. Please add you ssh public key to $HOME/.ssh/id_rsa.")
	// }

	// os.RemoveAll("./tmp")
	// _, err = git.PlainClone("./tmp/"+arr[1].Name, false, &git.CloneOptions{
	// 	URL:      "git@git.campmon.com:kenleyb/" + arr[1].Name + ".git",
	// 	Auth:     auth,
	// 	Progress: os.Stdout,
	// })
	// if err != nil {
	// 	fmt.Println("WDF", err)
	// 	return
	// }
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
