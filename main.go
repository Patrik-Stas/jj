package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func spotsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jjspots")
}

func loadSpots() map[string]string {
	spots := make(map[string]string)
	f, err := os.Open(spotsPath())
	if err != nil {
		return spots
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if parts := strings.SplitN(line, "\t", 2); len(parts) == 2 {
			spots[parts[0]] = parts[1]
		}
	}
	return spots
}

func saveSpots(spots map[string]string) error {
	names := make([]string, 0, len(spots))
	for k := range spots {
		names = append(names, k)
	}
	sort.Strings(names)

	f, err := os.Create(spotsPath())
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, name := range names {
		fmt.Fprintf(w, "%s\t%s\n", name, spots[name])
	}
	return w.Flush()
}

func cmdSave(args []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot get current directory")
		return 1
	}
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		name = filepath.Base(cwd)
		if name == "" || name == "/" || name == "." {
			fmt.Fprintln(os.Stderr, "error: cannot derive name from path, please provide one")
			return 1
		}
	}
	spots := loadSpots()
	spots[name] = cwd
	if err := saveSpots(spots); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("current directory %q saved as spot %q\n", cwd, name)
	return 0
}

func cmdRemove(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: jj remove <name>")
		return 1
	}
	name := args[0]
	spots := loadSpots()
	if _, ok := spots[name]; !ok {
		fmt.Fprintf(os.Stderr, "error: no spot named %q\n", name)
		return 1
	}
	delete(spots, name)
	if err := saveSpots(spots); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("removed spot %q\n", name)
	return 0
}

func cmdClean(args []string) int {
	spots := loadSpots()
	var removed []string
	for name, path := range spots {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			removed = append(removed, name)
			delete(spots, name)
		}
	}
	if len(removed) == 0 {
		fmt.Println("all spots are valid")
		return 0
	}
	sort.Strings(removed)
	if err := saveSpots(spots); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	for _, name := range removed {
		fmt.Printf("removed stale spot %q\n", name)
	}
	return 0
}

func cmdList(args []string) int {
	spots := loadSpots()
	prefix := ""
	if len(args) > 0 {
		prefix = args[0]
	}
	var matches []string
	for k := range spots {
		if strings.HasPrefix(k, prefix) {
			matches = append(matches, k)
		}
	}
	if len(matches) == 0 {
		return 0
	}
	sort.Strings(matches)
	maxLen := 0
	for _, name := range matches {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}
	for _, name := range matches {
		fmt.Printf("%-*s  %s\n", maxLen, name, spots[name])
	}
	return 0
}

func cmdResolve(args []string) int {
	if len(args) == 0 {
		return 1
	}
	name := args[0]
	spots := loadSpots()

	// exact match
	if path, ok := spots[name]; ok {
		fmt.Println(path)
		return 0
	}

	// prefix match
	var matches []string
	for k := range spots {
		if strings.HasPrefix(k, name) {
			matches = append(matches, k)
		}
	}
	if len(matches) == 1 {
		fmt.Println(spots[matches[0]])
		return 0
	}
	if len(matches) > 1 {
		sort.Strings(matches)
		fmt.Fprintf(os.Stderr, "error: ambiguous spot %q, matches: %s\n", name, strings.Join(matches, ", "))
		return 1
	}
	fmt.Fprintf(os.Stderr, "error: no spot named %q\n", name)
	return 1
}

func cmdComplete(args []string) int {
	prefix := ""
	if len(args) > 0 {
		prefix = args[0]
	}
	spots := loadSpots()
	var matches []string
	for k := range spots {
		if strings.HasPrefix(k, prefix) {
			matches = append(matches, k)
		}
	}
	sort.Strings(matches)
	for _, name := range matches {
		fmt.Println(name)
	}
	return 0
}

const shellZsh = `
# jj — jump jump: directory bookmarks
jj() {
    case "${1:-}" in
        ""|help|-h|--help)
            echo "jj — jump jump: directory bookmarks"
            echo ""
            echo "usage:"
            echo "  jj save [name]    save current directory as a spot (default: dirname)"
            echo "  jj list [prefix]  list spots"
            echo "  jj remove <name>  remove a spot"
            echo "  jj clean          remove spots with dead paths"
            echo "  jj <name>         jump to a spot"
            echo "  jj init <shell>   print shell setup code"
            ;;
        save|list|remove|clean|init)
            command _jj "$@"
            ;;
        *)
            local target
            target="$(command _jj resolve "$1" 2>/dev/null)"
            if [ $? -eq 0 ]; then
                cd "$target"
            else
                command _jj resolve "$1" >/dev/null
                return 1
            fi
            ;;
    esac
}

_jj_completions() {
    local -a spots
    spots=("${(@f)$(command _jj complete "${words[2]}" 2>/dev/null)}")
    if [[ ${#spots[@]} -gt 0 && -n "${spots[1]}" ]]; then
        compadd -a spots
    fi
}

compdef _jj_completions jj
`

const shellBash = `
# jj — jump jump: directory bookmarks
jj() {
    case "${1:-}" in
        ""|help|-h|--help)
            echo "jj — jump jump: directory bookmarks"
            echo ""
            echo "usage:"
            echo "  jj save [name]    save current directory as a spot (default: dirname)"
            echo "  jj list [prefix]  list spots"
            echo "  jj remove <name>  remove a spot"
            echo "  jj clean          remove spots with dead paths"
            echo "  jj <name>         jump to a spot"
            echo "  jj init <shell>   print shell setup code"
            ;;
        save|list|remove|clean|init)
            command _jj "$@"
            ;;
        *)
            local target
            target="$(command _jj resolve "$1" 2>/dev/null)"
            if [ $? -eq 0 ]; then
                cd "$target"
            else
                command _jj resolve "$1" >/dev/null
                return 1
            fi
            ;;
    esac
}

_jj_completions_bash() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"

    if [[ $COMP_CWORD -eq 1 ]]; then
        local spots
        spots="$(command _jj complete "$cur" 2>/dev/null)"
        local commands="save list remove init help"
        COMPREPLY=($(compgen -W "$spots $commands" -- "$cur"))
    elif [[ "$prev" == "remove" ]]; then
        local spots
        spots="$(command _jj complete "$cur" 2>/dev/null)"
        COMPREPLY=($(compgen -W "$spots" -- "$cur"))
    fi
}

complete -F _jj_completions_bash jj
`

func cmdInit(args []string) int {
	shell := "zsh"
	if len(args) > 0 {
		shell = args[0]
	}
	switch shell {
	case "zsh":
		fmt.Print(shellZsh)
	case "bash":
		fmt.Print(shellBash)
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported shell: %s\n", shell)
		return 1
	}
	return 0
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: _jj <command> [args]")
		os.Exit(1)
	}

	cmd := args[0]
	rest := args[1:]

	var code int
	switch cmd {
	case "save":
		code = cmdSave(rest)
	case "list":
		code = cmdList(rest)
	case "remove":
		code = cmdRemove(rest)
	case "clean":
		code = cmdClean(rest)
	case "resolve":
		code = cmdResolve(rest)
	case "complete":
		code = cmdComplete(rest)
	case "init":
		code = cmdInit(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		code = 1
	}
	os.Exit(code)
}
