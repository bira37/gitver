package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const (
	COLOR_GREEN  = "\x1b[32m"
	COLOR_RED    = "\x1b[31m"
	COLOR_YELLOW = "\x1b[33m"
	COLOR_END    = "\x1b[0m"
)

func create(dir string) {
	dir = strings.TrimSuffix(dir, "/")

	file, err := os.Create(fmt.Sprintf("%s/VERSION", dir))

	if err != nil {
		fmt.Printf("%sError creating file: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.Write([]byte("0.0.0"))

	if err != nil {
		fmt.Printf("%sError creating file: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%sCreated VERSION file at %s directory\n%s", COLOR_GREEN, dir, COLOR_END)
}

func increment(dir, label, inc, pre string) {
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

	// Check for valid increment mode
	if len(inc) > 0 && !slices.Contains([]string{"major", "minor", "patch"}, inc) {
		fmt.Printf("%sError: invalid value for increment mode\n%s", COLOR_RED, COLOR_END)
		os.Exit(1)
	}

	dir = strings.TrimSuffix(dir, "/")

	// Check for valid VERSION file
	content, err := os.ReadFile(fmt.Sprintf("%s/VERSION", dir))

	if err != nil {
		fmt.Printf("%sError reading file: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	// Read version
	ver, err := semver.NewVersion(string(content))

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

	// Set prerelease if applicable
	if len(pre) > 0 {
		*ver, err = ver.SetPrerelease(pre)

		if err != nil {
			fmt.Printf("%sError: invalid prerelease value: %v\n%s", COLOR_RED, err, COLOR_END)
			os.Exit(1)
		}
	}

	// Write to file
	err = os.WriteFile(fmt.Sprintf("%s/VERSION", dir), []byte(ver.String()), 0644)

	if err != nil {
		fmt.Printf("%sError writing to file: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%sUpdated VERSION to %s\n%s", COLOR_GREEN, ver.String(), COLOR_END)

	// Commit VERSION change
	prefix := ""

	if len(label) > 0 {
		prefix = fmt.Sprintf("%s-", label)
	}

	addCmd := exec.Command("git", "add", ".")
	addOut, err := addCmd.Output()

	if err != nil {
		fmt.Printf("%sError on git add: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%s%v\n%s", COLOR_YELLOW, string(addOut), COLOR_END)

	commitCmd := exec.Command("git", "commit", "-m", fmt.Sprintf(`"%s%s"`, prefix, ver.String()))
	commitOut, err := commitCmd.Output()

	if err != nil {
		fmt.Printf("%sError on git commit: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%s%v\n%s", COLOR_YELLOW, string(commitOut), COLOR_END)

	// Create tag
	tag := fmt.Sprintf("%s%s", prefix, ver.String())

	tagCmd := exec.Command("git", "tag", "-a", tag, "-m", fmt.Sprintf(`"%s"`, tag))
	_, err = tagCmd.Output()

	if err != nil {
		fmt.Printf("%sError on git tag: %v\n%s", COLOR_RED, err, COLOR_END)
		os.Exit(1)
	}

	fmt.Printf("%sSuccesfully tagged %s\n%s", COLOR_GREEN, tag, COLOR_END)
}

func main() {
	// CLI
	newFile := flag.String("new", "", "new file: creates a new file in specified directory with version <0.0.0>. Every other flag is ignored")

	dir := flag.String("d", "./", "directory: the path for the VERSION file")
	label := flag.String("l", "", "label: sets a prefix label only on the git tag (useful for monorepos to differentiate tags from multiple projects)")
	inc := flag.String("i", "", "increment mode: the increment type. valid inputs: major | minor | patch")
	pre := flag.String("p", "", "prerelease: sets a prerelease suffix")

	flag.Parse()

	// Execute
	if len(*newFile) > 0 {
		create(*newFile)
	} else {
		increment(*dir, *label, *inc, *pre)
	}
}
