package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/manifoldco/promptui"
)

type Config struct {
	AllowedLabels   []string `json:"allowedLabels"`
	AllowedBranches []string `json:"allowedBranches"`
}

const (
	COLOR_GREEN  = "\x1b[32m"
	COLOR_RED    = "\x1b[31m"
	COLOR_YELLOW = "\x1b[33m"
	COLOR_END    = "\x1b[0m"
)

func readConfig(configFile string) Config {
	file, err := os.ReadFile(configFile)
	mp := new(Config)

	if err != nil {
		return *mp
	}

	err = json.Unmarshal(file, mp)

	if err != nil {
		fmt.Printf("%sError reading config file (Malformed config json?): %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	return *mp
}

func getLatestTag(label string) string {
	// List all tags
	tags := make([]string, 0)
	if len(label) > 0 {
		// Remove label from tag
		tagLabelCmd := exec.Command("git", "--no-pager", "tag", "-l", fmt.Sprintf("%s-*", label))
		tagLabelOutput, err := tagLabelCmd.Output()
		if err != nil {
			fmt.Printf("%sError on git tag list: %v\n%s", COLOR_RED, err, COLOR_END)
			os.Exit(1)
		}
		rawTags := strings.Split(string(tagLabelOutput), "\n")

		for _, rawTag := range rawTags {
			tag := strings.Replace(rawTag, fmt.Sprintf("%s-", label), "", 1)
			tag = strings.TrimSpace(tag)
			if len(tag) == 0 {
				continue
			}
			tags = append(tags, tag)
		}
	} else {
		// Do not need to remove label from tag
		tagCmd := exec.Command("git", "--no-pager", "tag")
		tagOutput, err := tagCmd.Output()
		if err != nil {
			fmt.Printf("%sError on git tag list: %v\n%s", COLOR_RED, err, COLOR_END)
			os.Exit(1)
		}
		rawTags := strings.Split(string(tagOutput), "\n")
		for _, rawTag := range rawTags {
			tag := strings.TrimSpace(rawTag)
			if len(tag) == 0 {
				continue
			}
			tags = append(tags, tag)
		}
	}

	// Get the latest tag
	sort.Slice(tags, func(i, j int) bool {
		a, err := semver.NewVersion(tags[i])
		b, err2 := semver.NewVersion(tags[j])

		if err != nil || err2 != nil {
			fmt.Printf("%sError: tags in this project does not follow SemVer format\n%s", COLOR_RED, COLOR_END)
			os.Exit(1)
		}

		return a.Compare(b) >= 0
	})

	return tags[0]
}

func yesNoPrompt(label string) bool {
	prompt := promptui.Select{
		Label: fmt.Sprintf("%s", label),
		Items: []string{"no", "yes"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("%sError: prompt failed:%v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}
	return result == "yes"
}

func increment(label, inc, pre string, release bool, configFile string) {
	// Get the config
	config := readConfig(configFile)

	// Check label restrictions
	if config.AllowedLabels != nil {
		// Check if label is required to exist, and then check if exists
		if len(label) == 0 {
			fmt.Printf("%sError: A label is required to exist to create the tag\n%s", COLOR_RED, COLOR_END)
			os.Exit(1)
		}

		// Check if label is included on allowed labels
		if !slices.Contains(config.AllowedLabels, label) {
			fmt.Printf("%sError: The label is not allowed to be used\n%s", COLOR_RED, COLOR_END)
			os.Exit(1)
		}
	}

	// Check if there are branch restrictions
	if config.AllowedBranches != nil {
		currentBranchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		currentBranch, err := currentBranchCmd.Output()

		if err != nil {
			fmt.Printf("%sError: unable to get current branch: %v\n%s", COLOR_RED, err, COLOR_END)
			os.Exit(1)
		}

		if !slices.Contains(config.AllowedBranches, strings.TrimSpace(string(currentBranch))) {
			fmt.Println(strings.TrimSpace(string(currentBranch)))
			fmt.Printf("%sError: current branch is not allowed to be tagged\n%s", COLOR_RED, COLOR_END)
			os.Exit(1)
		}
	}

	// Check if Git working directory is clean
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOut, err := statusCmd.Output()

	if err != nil {
		fmt.Printf("%sError executing git command: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	if len(string(statusOut)) != 0 {
		fmt.Printf("%sError: Git working directory not clean. Commit your changes before versioning.\n%s", COLOR_RED, COLOR_END)
		os.Exit(1)
	}

	// Pull git changes and tags
	pullCmd := exec.Command("git", "pull", "--tags")
	_, err = pullCmd.Output()

	if err != nil {
		fmt.Printf("%sError executing git pull command: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%sPulled remote changes and tags\n%s", COLOR_GREEN, COLOR_END)

	// Get the latest tag
	latestTag := getLatestTag(label)

	// Check for valid increment mode
	if len(inc) > 0 && !slices.Contains([]string{"major", "minor", "patch"}, inc) {
		fmt.Printf("%sError: invalid value for increment mode\n%s", COLOR_RED, COLOR_END)
		os.Exit(1)
	}

	// Read version
	ver, err := semver.NewVersion(latestTag)

	if err != nil {
		fmt.Printf("%sError: could not parse version: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	// Increment version
	switch inc {
	case "major":
		*ver = ver.IncMajor()
	case "minor":
		*ver = ver.IncMinor()
	case "patch":
		*ver = ver.IncPatch()
	default:
	}

	// Set prerelease
	if len(pre) > 0 || release {
		suf := pre

		if release {
			suf = ""
		}

		*ver, err = ver.SetPrerelease(suf)

		if err != nil {
			fmt.Printf("%sError: invalid prerelease value: %v\n%s", COLOR_RED, err, COLOR_END)
			os.Exit(1)
		}
	}

	// Create tag
	prefix := ""

	if len(label) > 0 {
		prefix = fmt.Sprintf("%s-", label)
	}
	tag := fmt.Sprintf("%s%s", prefix, ver.String())

	if !yesNoPrompt(fmt.Sprintf("Creating tag <%s>. Are you sure?", tag)) {
		os.Exit(0)
	}

	tagCmd := exec.Command("git", "tag", "-a", tag, "-m", fmt.Sprintf(`"%s"`, tag))
	_, err = tagCmd.Output()

	if err != nil {
		fmt.Printf("%sError on git tag (maybe tag already exists?): %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%sSuccesfully tagged %s\n%s", COLOR_GREEN, tag, COLOR_END)

	// Push tag
	pushCmd := exec.Command("git", "push", "--tags")
	_, err = pushCmd.Output()

	if err != nil {
		fmt.Printf("%sError pushing tags: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%sSuccesfully pushed tag %s\n%s", COLOR_GREEN, tag, COLOR_END)
}

func main() {
	// CLI
	label := flag.String("l", "", "label: sets a prefix label only on the git tag (useful for monorepos to differentiate tags from multiple projects)")
	inc := flag.String("i", "", "increment mode: the increment type. valid inputs: major | minor | patch")
	pre := flag.String("p", "", "prerelease: sets a prerelease suffix")
	release := flag.Bool("r", false, "release: removes prerelease suffix. Overrides prerelease option")
	configFile := flag.String("config", "./gitver.json", "config: location of the config file. If not provided, try to find in current dir. Config file has priority over directory flag")

	flag.Parse()

	// Execute
	increment(*label, *inc, *pre, *release, *configFile)
}
