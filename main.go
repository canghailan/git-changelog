package main

import (
	"bufio"
	"fmt"
	"github.com/atotto/clipboard"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Options struct {
	Repo        string
	FromVersion string
	ToVersion   string
}

func main() {
	options := Input()
	logs, err := GitLog(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "执行git log失败\n")
		log.Fatal(err)
	}
	changelogs := strings.Join(Format(logs), "\n")
	Output(changelogs)
	Wait()
}

func Input() Options {
	fmt.Print("    Git Repo: ")
	repo := strings.TrimSpace(ReadClipboard())
	fmt.Println(repo)

	fmt.Print("From Version: ")
	fromVersion := strings.TrimSpace(ReadClipboard())
	fmt.Println(fromVersion)

	fmt.Print("  To Version: ")
	toVersion := strings.TrimSpace(ReadClipboard())
	fmt.Println(toVersion)

	return Options{repo, fromVersion, toVersion}
}

func ReadStdin() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}

func ReadClipboard() string {
	clipboard.WriteAll("")
	for {
		text, _ := clipboard.ReadAll()
		if len(text) != 0 {
			return text
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func GitLog(options Options) (string, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%s", "--reverse", options.FromVersion+".."+options.ToVersion)
	if len(options.Repo) != 0 {
		cmd.Dir = options.Repo
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return "", err
	}
	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func Format(logs string) []string {
	lines := strings.Split(string(logs), "\n")

	index := 0
	filter := map[string]bool{}
	changelogs := []string{}

	for _, line := range lines {
		if isChangelog, _ := regexp.MatchString("^\\s*[0-9].*$", line); isChangelog {
			changelog_list := regexp.MustCompile("[0-9]+\\.").Split(line, -1)
			for _, changelog := range changelog_list {
				changelog = regexp.MustCompile("(^[\\s0-9]+|\\s+$)").ReplaceAllString(changelog, "")
				if len(changelog) == 0 {
					continue
				}
				if isTemp, _ := regexp.MatchString("^临时", changelog); isTemp {
					continue
				}
				if filter[changelog] {
					continue
				}
				filter[changelog] = true

				index = index + 1
				changelogs = append(changelogs, fmt.Sprintf("%3d. %s", index, changelog))
			}
		}
	}

	return changelogs
}

func Output(result string) {
	fmt.Println("\nCHANGELOG:")
	fmt.Println(result)
	clipboard.WriteAll(result)
}

func Wait() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
