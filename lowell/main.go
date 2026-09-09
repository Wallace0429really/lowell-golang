package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// GLOBAL SYMBOL TABLE & STATE
var vars = make(map[string]string)
var shortcuts = make(map[string]string) // alias -> command
var history []string                    // Command history for REPL

func main() {
	if len(os.Args) < 2 {
		fmt.Println("          <L>")
		fmt.Println("  Welcome to Lowell Stable 1.0!")
		fmt.Println("  Type 'lowell repl' to start interactive mode.")
		fmt.Println("  Type 'lowell run file.low' to run scripts.")
		fmt.Println("==========================================")
		return
	}
	command := resolveAlias(os.Args[1])
	switch command {
	case "repl":
		startREPL()
	case "run":
		handleRunCLI(os.Args[2:])
	case "print":
		if len(os.Args) < 3 {
			fmt.Println("[ERROR] Usage: lowell print <expression>")
			return
		}
		arg := strings.Join(os.Args[2:], " ")
		executePrint(arg, true)
	case "exec":
		if len(os.Args) < 3 {
			fmt.Println("[ERROR] Usage: lowell exec <command>")
			return
		}
		cmd := strings.Join(os.Args[2:], " ")
		handleExecCommand(cmd, true)
	case "help":
		if len(os.Args) < 3 {
			printHelpIndex()
		} else {
			searchHelp(os.Args[2])
		}
	default:
		fmt.Printf("Unknown command '%s'. Try 'lowell help'.\n", command)
	}
}

// ✅ ALIAS RESOLUTION ENGINE (CLI single-token resolve)
func resolveAlias(cmd string) string {
	if alias, exists := shortcuts[cmd]; exists {
		return alias
	}
	return cmd
}

// ✅ Whole line alias expand for REPL: replace first token if matches alias
func expandWholeLine(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return line
	}
	first := fields[0]
	if repl, ok := shortcuts[first]; ok {
		return strings.Replace(line, first, repl, 1)
	}
	return line
}

// ✅ HELPER: GET BASE COMMAND (handles flags like --kwd)
func getBaseCommand(line string) string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}
	base := parts[0]
	// Check if base is an alias that maps to a multi-word command
	if resolved := resolveAlias(base); resolved != base {
		resolvedParts := strings.Fields(resolved)
		if len(resolvedParts) > 0 {
			return resolvedParts[0]
		}
	}
	return base
}

// ✅ THE REPL ENGINE (V1.0 - FULL SUITE + SMART NAV!)
func startREPL() {
	fmt.Println("Lowell Stable 1.0 Interactive REPL")
	fmt.Println("Commands: print, let, answer, vars, del/dr, edit/et, run/riso,")
	fmt.Println("          history/hk, clear/cc, navigate/nvt, shortcut/sclist, help, exit")
	fmt.Println("Type 'exit' or 'quit' to leave. Variables persist!")
	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "?"
		}
		fmt.Printf("%s lowell> ", cwd)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "exit" || line == "quit" {
			fmt.Println("Goodbye! 👋")
			break
		}
		// Add to history before alias expansion (store user original input)
		history = append(history, line)

		// ✅ Expand alias whole-line BEFORE parsing command
		line = expandWholeLine(line)

		lowerLine := strings.ToLower(line)
		resolvedCmd := getBaseCommand(lowerLine)
		switch resolvedCmd {
		case "print":
			executePrint(strings.TrimPrefix(line, getBaseCommand(line)), true)
		case "let":
			executeDef(line, true)
		case "answer":
			executeAnswer(line, true)
		case "vars":
			listVars()
		case "del", "dr":
			handleDel(line, lowerLine, true)
		case "edit", "et":
			handleEdit(strings.TrimSpace(strings.TrimPrefix(line, getBaseCommand(line))), true)
		case "run":
			handleRunREPL(strings.TrimSpace(strings.TrimPrefix(line, getBaseCommand(line))), false)
		case "riso":
			handleRunREPL(strings.TrimSpace(strings.TrimPrefix(line, getBaseCommand(line))), true)
		case "history", "hk":
			handleHistory(line, lowerLine, true)
		case "clear", "cc":
			handleClear(line, lowerLine, true)
		case "shortcut":
			handleShortcut(line, lowerLine, true)
		case "sclist", "shortcutlist":
			listShortcuts()
		case "navigate", "nvt":
			handleNavigate(line, true)
		case "if":
			fmt.Println("[NOTE] Multi-line blocks (if/repeat) require .low files for now!")
		case "help":
			topic := strings.TrimSpace(strings.TrimPrefix(line, getBaseCommand(line)))
			if topic == "" {
				printHelpIndex()
			} else {
				if !strings.HasPrefix(topic, "-") {
					topic = "-" + topic
				}
				searchHelp(topic)
			}
		default:
			fmt.Println("[ERROR] Unknown REPL command. Type 'help' or 'sclist' for guidance.")
		}
	}
}

// ✅ SMART NAVIGATE LOCATION
func handleNavigate(line string, interactive bool) {
	target := strings.TrimSpace(strings.TrimPrefix(line, getBaseCommand(line)))

	// Bare 'navigate' goes home
	if target == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("[ERROR] Cannot determine home directory.")
		} else {
			os.Chdir(home)
			if interactive {
				fmt.Printf("[OK] Navigated to home: %s\n", home)
			}
		}
		return
	}

	// Resolve path
	resolved := target
	if !filepath.IsAbs(target) {
		cwd, _ := os.Getwd()
		resolved = filepath.Join(cwd, target)
	}

	// Validate directory
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		if interactive {
			fmt.Printf("[ERROR] '%s' is not a valid location.\n", target)
			// Smart suggestions
			parent := filepath.Dir(resolved)
			if entries, readErr := os.ReadDir(parent); readErr == nil {
				var matches []string
				searchName := strings.ToLower(filepath.Base(target))
				for _, e := range entries {
					if e.IsDir() && strings.HasPrefix(strings.ToLower(e.Name()), searchName) {
						matches = append(matches, e.Name())
					}
				}
				if len(matches) > 0 {
					fmt.Printf("💡 Did you mean: %v?\n", matches)
				}
			}
		}
		return
	}

	// Change directory
	if err := os.Chdir(resolved); err != nil {
		if interactive {
			fmt.Printf("[ERROR] Permission denied: %s\n", resolved)
		}
	} else if interactive {
		newCwd, _ := os.Getwd()
		fmt.Printf("[OK] Navigated to: %s\n", newCwd)
	}
}

// ✅ LIST ALL CURRENT VARIABLES
func listVars() {
	if len(vars) == 0 {
		fmt.Println("No variables defined yet.")
		return
	}
	fmt.Println("Current Variables:")
	i := 1
	for k, v := range vars {
		fmt.Printf("  %d. $%s = \"%s\"\n", i, k, v)
		i++
	}
}

// ✅ LIST ALL SHORTCUTS
func listShortcuts() {
	if len(shortcuts) == 0 {
		fmt.Println("No custom shortcuts defined.")
		return
	}
	fmt.Println("Active Shortcuts:")
	for alias, cmd := range shortcuts {
		fmt.Printf("  %-8s → %s\n", alias, cmd)
	}
}

// ✅ DELETE VARIABLES (Multi-mode: explicit, wildcard, index)
func handleDel(line, lowerLine string, interactive bool) {
	args := strings.Fields(strings.TrimPrefix(line, getBaseCommand(line)))
	if len(args) == 0 {
		fmt.Println("[ERROR] Usage: del $name | del $prefix* | del N-M")
		return
	}
	deletedCount := 0
	for _, arg := range args {
		// Index range deletion (e.g., "1-3")
		if strings.Contains(arg, "-") && !strings.HasPrefix(arg, "$") {
			parts := strings.Split(arg, "-")
			if len(parts) == 2 {
				start, err1 := strconv.Atoi(parts[0])
				end, err2 := strconv.Atoi(parts[1])
				if err1 == nil && err2 == nil && start >= 1 && end >= start {
					keys := getVarKeys()
					for idx := start; idx <= end && idx <= len(keys); idx++ {
						delete(vars, keys[idx-1])
						deletedCount++
					}
					continue
				}
			}
		}
		// Wildcard deletion (e.g., "$temp*")
		if strings.HasSuffix(arg, "*") && strings.HasPrefix(arg, "$") {
			prefix := arg[1 : len(arg)-1]
			for k := range vars {
				if strings.HasPrefix(k, prefix) {
					delete(vars, k)
					deletedCount++
				}
			}
			continue
		}
		// Explicit deletion
		name := arg
		if !strings.HasPrefix(name, "$") {
			name = "$" + name
		}
		key := name[1:]
		if _, exists := vars[key]; exists {
			delete(vars, key)
			deletedCount++
		} else if interactive {
			fmt.Printf("[WARN] \"$%s\" not found. Skipping.\n", key)
		}
	}
	if interactive {
		fmt.Printf("[OK] Deleted %d variable(s).\n", deletedCount)
	}
}

func getVarKeys() []string {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	return keys
}

// ✅ EDIT COMMAND (Non-blocking, auto-create, fallback editor)
func handleEdit(filename string, interactive bool) {
	if filename == "" {
		fmt.Println("[ERROR] Usage: edit <filename.low>")
		return
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".low") {
		filename += ".low"
	}
	// Create blank file if missing
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.WriteFile(filename, []byte(""), 0644)
		if interactive {
			fmt.Printf("[OK] Created blank '%s'.\n", filename)
		}
	}
	// Open with system editor (non-blocking)
	var cmd *exec.Cmd
	editor := os.Getenv("EDITOR")
	if editor != "" {
		cmd = exec.Command(editor, filename)
	} else if runtime.GOOS == "windows" {
		// Try VS Code first, then Notepad
		if _, err := exec.LookPath("code"); err == nil {
			cmd = exec.Command("code", filename)
		} else {
			cmd = exec.Command("notepad", filename)
		}
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", filename)
	} else {
		cmd = exec.Command("xdg-open", filename)
	}
	if cmd != nil {
		err := cmd.Start()
		if err != nil {
			fmt.Printf("[ERROR] Could not open editor: %v\n", err)
			return
		}
		if interactive {
			fmt.Printf("[OK] Opening '%s' in editor...\n", filename)
			fmt.Println(" Tip: Save in editor, then type 'run "+filename+"' to test!")
		}
	}
}

// ✅ RUN COMMAND HANDLER (Persistent vs Isolated)
func handleRunCLI(args []string) {
	if len(args) == 0 {
		fmt.Println("[ERROR] Usage: lowell run <file.low> | lowell run --iso <file.low>")
		return
	}
	isolated := false
	filename := args[0]
	if args[0] == "--iso" && len(args) > 1 {
		isolated = true
		filename = args[1]
	}
	runFileInternal(filename, isolated, false)
}

func handleRunREPL(arg string, isolated bool) {
	if arg == "" {
		fmt.Println("[ERROR] Usage: run <file.low> | riso <file.low>")
		return
	}
	runFileInternal(arg, isolated, true)
}

func runFileInternal(filename string, isolated bool, interactive bool) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("[ERROR] Cannot read '%s': %v\n", filename, err)
		return
	}
	lines := strings.Split(string(data), "\n")

	// Isolated mode uses temporary var map
	var execVars map[string]string
	if isolated {
		execVars = make(map[string]string)
	} else {
		execVars = vars
	}
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}
		lowerLine := strings.ToLower(line)
		if strings.HasPrefix(lowerLine, "if ") {
			i = handleIfBlockWithVars(lines, i, execVars)
			continue
		}
		if strings.HasPrefix(lowerLine, "repeat ") {
			i = handleRepeatBlockWithVars(lines, i, execVars)
			continue
		}
		if strings.HasPrefix(lowerLine, "answer ") {
			if !isolated {
				executeAnswerWithVars(line, false, execVars)
			} else if interactive {
				fmt.Println("[WARN] 'answer' skipped in isolated mode.")
			}
			i++
			continue
		}
		if strings.HasPrefix(lowerLine, "let ") {
			executeDefWithVars(line, false, execVars)
		} else if strings.HasPrefix(lowerLine, "print ") {
			executePrintWithVars(strings.TrimPrefix(line, "print "), false, execVars)
		} else if !interactive {
			fmt.Printf("[ERROR] Line %d: Unknown command.\n", i+1)
		}
		i++
	}
	if interactive {
		mode := "persistent"
		if isolated {
			mode = "isolated"
		}
		fmt.Printf("[OK] Executed '%s' in %s mode.\n", filename, mode)
	}
}

// ✅ HISTORY HANDLER (Search, Export, Recent)
func handleHistory(line, lowerLine string, interactive bool) {
	args := strings.Fields(strings.TrimPrefix(line, getBaseCommand(line)))

	if len(args) == 0 {
		// Show all history
		for _, h := range history {
			fmt.Println(h)
		}
		return
	}
	// Numeric: show last N
	if n, err := strconv.Atoi(args[0]); err == nil {
		start := len(history) - n
		if start < 0 {
			start = 0
		}
		for _, h := range history[start:] {
			fmt.Println(h)
		}
		return
	}
	// Keyword search
	if args[0] == "--kwd" || args[0] == "--keyword" {
		if len(args) < 2 {
			fmt.Println("[ERROR] Usage: history --kwd \"term\"")
			return
		}
		term := strings.ToLower(strings.Join(args[1:], " "))
		found := false
		for _, h := range history {
			if strings.Contains(strings.ToLower(h), term) {
				fmt.Println(h)
				found = true
			}
		}
		if !found && interactive {
			fmt.Println("[INFO] No matching commands found.")
		}
		return
	}
	// Download/Export
	if args[0] == "--dwd" || args[0] == "--download" {
		if len(args) < 2 {
			fmt.Println("[ERROR] Usage: history --dwd <filename.txt>")
			return
		}
		fname := args[1]
		if !strings.HasSuffix(strings.ToLower(fname), ".txt") {
			fname += ".txt"
		}
		content := strings.Join(history, "\n")
		err := os.WriteFile(fname, []byte(content), 0644)
		if err != nil {
			fmt.Printf("[ERROR] Failed to save '%s': %v\n", fname, err)
		} else if interactive {
			fmt.Printf("[OK] Saved %d commands to '%s'.\n", len(history), fname)
		}
		return
	}
}

// ✅ CLEAR FAMILY HANDLER (Surgical State Management)
func handleClear(line, lowerLine string, interactive bool) {
	args := strings.Fields(strings.TrimPrefix(line, getBaseCommand(line)))

	if len(args) == 0 {
		// Bare clear requires confirmation
		if interactive {
			fmt.Println("[WARN] This will delete ALL variables and reset the session!")
			fmt.Println("      Type 'clear --confirm' to proceed, or use:")
			fmt.Println("        clear --keepvars / --kvs  → Clear screen only")
			fmt.Println("        clear --onlyvars / --ovs  → Wipe vars, keep aliases")
			fmt.Println("        clear --shortcuts / --scs → Wipe aliases, keep vars")
		}
		return
	}
	flag := args[0]
	switch flag {
	case "--confirm", "--cm":
		vars = make(map[string]string)
		history = nil
		if interactive {
			fmt.Println("[OK] Full session reset complete.")
		}
	case "--keepvars", "--kvs":
		// Screen/history clear only (simulated by printing separator)
		fmt.Println("\n--- Screen Cleared ---")
		if interactive {
			fmt.Println("[OK] Screen cleared. Variables preserved.")
		}
	case "--onlyvars", "--ovs":
		vars = make(map[string]string)
		if interactive {
			fmt.Println("[OK] Variables cleared. Aliases preserved.")
		}
	case "--shortcuts", "--scs":
		shortcuts = make(map[string]string)
		if interactive {
			fmt.Println("[OK] Custom shortcuts removed. Variables intact.")
		}
	default:
		if interactive {
			fmt.Printf("[ERROR] Unknown clear flag '%s'. Type 'help -clear'.\n", flag)
		}
	}
}

// ✅ SHORTCUT SYSTEM (Natural Language Aliases)
func handleShortcut(line, lowerLine string, interactive bool) {
	// Parse "shortcut "<cmd>" into "<alias>"
	intoIdx := strings.Index(lowerLine, " into ")
	if intoIdx == -1 {
		// Check for delete command
		args := strings.Fields(strings.TrimPrefix(line, getBaseCommand(line)))
		if len(args) >= 2 && args[0] == "delete" {
			aliasName := args[1]
			if _, exists := shortcuts[aliasName]; exists {
				delete(shortcuts, aliasName)
				if interactive {
					fmt.Printf("[OK] Alias '%s' deleted.\n", aliasName)
				}
			} else if interactive {
				fmt.Printf("[WARN] Alias '%s' not found.\n", aliasName)
			}
			return
		}

		if interactive {
			fmt.Println("[ERROR] Usage: shortcut \"<cmd>\" into \"<alias>\" | shortcut delete <alias>")
			fmt.Println("      Type 'sclist' to see all active shortcuts.")
		}
		return
	}
	cmdPart := strings.Trim(strings.TrimSpace(line[:intoIdx]), "\"")
	// Remove "shortcut " prefix from cmdPart, NO ToLower(), preserve original case
	cmdPart = strings.TrimPrefix(cmdPart, "shortcut ")
	cmdPart = strings.TrimSpace(cmdPart)

	aliasPart := strings.Trim(strings.TrimSpace(line[intoIdx+6:]), "\"")

	if cmdPart == "" || aliasPart == "" {
		fmt.Println("[ERROR] Both command and alias must be specified.")
		return
	}

	shortcuts[aliasPart] = cmdPart
	if interactive {
		fmt.Printf("[OK] Alias '%s' created for '%s'.\n", aliasPart, cmdPart)
	}
}

// ✅ HELP SYSTEM (Modular Topics)
func searchHelp(topic string) {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("[ERROR] Cannot locate Lowell installation.")
		return
	}
	exeDir := filepath.Dir(exePath)
	helpPath := filepath.Join(exeDir, "HELP.md")
	data, err := os.ReadFile(helpPath)
	if err != nil {
		fmt.Printf("[ERROR] HELP.md not found at: %s\n", helpPath)
		return
	}
	content := string(data)
	topicTag := fmt.Sprintf("# [%s]", strings.ToUpper(strings.TrimPrefix(topic, "-")))
	if strings.Contains(content, topicTag) {
		parts := strings.Split(content, topicTag)
		if len(parts) > 1 {
			nextSection := strings.Split(parts[1], "# [")[0]
			fmt.Println(strings.TrimSpace(nextSection))
		}
	} else {
		fmt.Printf("[ERROR] Topic '%s' not found. Type 'help' for available topics.\n", topic)
	}
}

func printHelpIndex() {
	fmt.Println("Available Help Topics:")
	fmt.Println("  help -general    → Core syntax & basics")
	fmt.Println("  help -math       → Angle bracket computation")
	fmt.Println("  help -vars       → Variables & deletion modes")
	fmt.Println("  help -logic      → If/else decision blocks")
	fmt.Println("  help -loops      → Repeat actions easily")
	fmt.Println("  help -input      → User input handling")
	fmt.Println("  help -navigation → Directory navigation (navigate/nvt)")
	fmt.Println("  help -edit       → File creation & editing workflow")
	fmt.Println("  help -run        → Persistent & isolated execution")
	fmt.Println("  help -clear      → Surgical state management")
	fmt.Println("  help -history    → Searchable command memory")
	fmt.Println("  help -shortcuts  → Natural language aliases")
	fmt.Println("  help -errors     → Smart error messages guide")
}

// ✅ SEMANTIC ENGINES (With Variable Map Support)
func executePrint(arg string, interactive bool) {
	executePrintWithVars(arg, interactive, vars)
}

func executePrintWithVars(arg string, interactive bool, v map[string]string) {
	arg = strings.TrimSpace(arg)
	if strings.HasPrefix(arg, "<") && strings.HasSuffix(arg, ">") {
		fmt.Println(evaluateMathWithVars(arg[1:len(arg)-1], v))
	} else if strings.HasPrefix(arg, `"`) && strings.HasSuffix(arg, `"`) && len(arg) >= 2 {
		fmt.Println(arg[1 : len(arg)-1])
	} else if strings.Contains(arg, "(") || strings.Contains(arg, ")") {
		inner := strings.Trim(arg, "()")
		fmt.Println(evaluateGlueWithVars(inner, v))
	} else if strings.HasPrefix(arg, "$") {
		key := arg[1:]
		if val, exists := v[key]; exists {
			fmt.Println(val)
		} else {
			msg := fmt.Sprintf("[ERROR] \"$%s\" is not defined!", key)
			if interactive {
				msg += fmt.Sprintf(" Use \"let $%s be <val>\" to define it!", key)
			}
			fmt.Println(msg)
		}
	} else {
		msg := "[ERROR] Invalid syntax! Variables must start with \"$\"."
		if interactive {
			msg += fmt.Sprintf(" Did you mean \"$%s\"?", arg)
		}
		fmt.Println(msg)
	}
}

func evaluateMath(expr string) string {
	return evaluateMathWithVars(expr, vars)
}

func evaluateMathWithVars(expr string, v map[string]string) string {
	expr = strings.ReplaceAll(expr, " ", "")
	opIdx := -1
	opChar := ""
	for i, r := range expr {
		if r == '+' || r == '-' || r == '*' || r == '/' {
			if i == 0 || expr[i-1] == '+' || expr[i-1] == '-' || expr[i-1] == '*' || expr[i-1] == '/' {
				continue
			}
			opIdx = i
			opChar = string(r)
			break
		}
	}
	if opIdx == -1 {
		val, err := resolveOperandWithVars(expr, v)
		if err != nil {
			return fmt.Sprintf("[ERROR] %v", err)
		}
		return strconv.Itoa(val)
	}
	left := expr[:opIdx]
	right := expr[opIdx+1:]
	a, errA := resolveOperandWithVars(left, v)
	b, errB := resolveOperandWithVars(right, v)
	if errA != nil {
		return fmt.Sprintf("[ERROR] Left operand: %v", errA)
	}
	if errB != nil {
		return fmt.Sprintf("[ERROR] Right operand: %v", errB)
	}
	switch opChar {
	case "+":
		return strconv.Itoa(a + b)
	case "-":
		return strconv.Itoa(a - b)
	case "*":
		return strconv.Itoa(a * b)
	case "/":
		if b == 0 {
			return "[ERROR] Division by zero"
		}
		return strconv.Itoa(a / b)
	default:
		return "[ERROR] Unknown operator"
	}
}

func resolveOperand(s string) (int, error) {
	return resolveOperandWithVars(s, vars)
}

func resolveOperandWithVars(s string, v map[string]string) (int, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "$") {
		key := s[1:]
		if val, exists := v[key]; exists {
			n, err := strconv.Atoi(val)
			if err != nil {
				return 0, fmt.Errorf("var \"$%s\" is text (\"%s\"), not a number", key, val)
			}
			return n, nil
		}
		return 0, fmt.Errorf("undefined variable: %s", key)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("non-numeric literal: \"%s\"", s)
	}
	return n, nil
}

func evaluateGlue(expr string) string {
	return evaluateGlueWithVars(expr, vars)
}

func evaluateGlueWithVars(expr string, v map[string]string) string {
	segments := strings.Split(expr, "+")
	var output strings.Builder
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, `"`) && strings.HasSuffix(seg, `"`) && len(seg) >= 2 {
			output.WriteString(seg[1 : len(seg)-1])
		} else if strings.HasPrefix(seg, "$") {
			if val, exists := v[seg[1:]]; exists {
				output.WriteString(val)
			}
		} else {
			output.WriteString(seg)
		}
	}
	return output.String()
}

func executeDef(cmd string, interactive bool) {
	executeDefWithVars(cmd, interactive, vars)
}

func executeDefWithVars(cmd string, interactive bool, v map[string]string) {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(strings.ToLower(cmd), "let ") {
		msg := `[ERROR] Lowell uses "let...be".`
		if interactive {
			msg += ` Try: let $x be <5>`
		}
		fmt.Println(msg)
		return
	}
	rest := strings.TrimSpace(cmd[4:])
	beIdx := strings.Index(strings.ToLower(rest), " be ")
	if beIdx == -1 {
		fmt.Println("[ERROR] Missing \"be\" keyword!")
		return
	}
	varPart := strings.TrimSpace(rest[:beIdx])
	rhs := strings.TrimSpace(rest[beIdx+4:])
	if rhs == "" {
		fmt.Println("[ERROR] Did you forget the value? :3")
		return
	}
	if !strings.HasPrefix(varPart, "$") {
		fmt.Println("[ERROR] Variable names must start with \"$\"!")
		return
	}
	key := varPart[1:]
	if !isValidVarName(key) {
		fmt.Printf("[ERROR] Invalid character in \"$%s\".\n", key)
		return
	}
	if _, exists := v[key]; exists {
		if interactive {
			fmt.Printf("[WARNING] Redefining \"$%s\".\n", key)
		}
	}
	var value string
	if strings.HasPrefix(rhs, "<") && strings.HasSuffix(rhs, ">") {
		value = evaluateMathWithVars(rhs[1:len(rhs)-1], v)
		if strings.HasPrefix(value, "[ERROR]") {
			fmt.Printf("[ERROR] In definition: %s\n", value)
			return
		}
	} else if strings.HasPrefix(rhs, `"`) && strings.HasSuffix(rhs, `"`) && len(rhs) >= 2 {
		value = rhs[1 : len(rhs)-1]
	} else if strings.Contains(rhs, "(") || strings.Contains(rhs, ")") {
		value = evaluateGlueWithVars(strings.Trim(rhs, "()"), v)
	} else if strings.HasPrefix(rhs, "$") {
		if val, exists := v[rhs[1:]]; exists {
			value = val
		} else {
			return
		}
	} else {
		value = rhs
	}
	v[key] = value
}

func executeAnswer(cmd string, interactive bool) {
	executeAnswerWithVars(cmd, interactive, vars)
}

func executeAnswerWithVars(cmd string, interactive bool, v map[string]string) {
	cmd = strings.TrimSpace(cmd)
	rest := strings.TrimPrefix(strings.ToLower(cmd), "answer ")
	beIdx := strings.Index(rest, " be ")
	if beIdx == -1 {
		fmt.Println("[ERROR] Invalid answer syntax! Use: answer $var be \"prompt\"")
		return
	}
	varPart := strings.TrimSpace(rest[:beIdx])
	promptPart := strings.TrimSpace(rest[beIdx+4:])
	if !strings.HasPrefix(varPart, "$") {
		fmt.Println("[ERROR] Variable must start with '$'!")
		return
	}
	key := varPart[1:]
	if !isValidVarName(key) {
		fmt.Printf("[ERROR] Invalid variable name \"$%s\".\n", key)
		return
	}
	expectNumber := false
	promptText := ""
	if strings.HasPrefix(promptPart, "<") && strings.HasSuffix(promptPart, ">") {
		expectNumber = true
		inner := promptPart[1 : len(promptPart)-1]
		if _, err := strconv.Atoi(inner); err != nil {
			promptText = inner
		} else {
			promptText = "Enter a number: "
		}
	} else if strings.HasPrefix(promptPart, `"`) && strings.HasSuffix(promptPart, `"`) && len(promptPart) >= 2 {
		promptText = promptPart[1 : len(promptPart)-1]
	} else {
		promptText = promptPart
	}
	fmt.Print(promptText)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("\n[ERROR] Failed to read input.")
		return
	}
	input = strings.TrimSpace(input)
	if expectNumber {
		_, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("[ERROR] Expected a number for \"$%s\", but got \"%s\".\n", key, input)
			return
		}
	}
	v[key] = input
	if interactive {
		fmt.Printf("[DEBUG] Stored \"$%s\" = \"%s\"\n", key, input)
	}
}

func isValidVarName(name string) bool {
	if len(name) == 0 {
		return false
	}
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

// ✅ LOGIC ENGINES (With Variable Map Support)
func handleIfBlock(lines []string, startIndex int) int {
	return handleIfBlockWithVars(lines, startIndex, vars)
}

func handleIfBlockWithVars(lines []string, startIndex int, v map[string]string) int {
	line := strings.TrimSpace(lines[startIndex])
	content := strings.TrimPrefix(strings.ToLower(line), "if ")
	content = strings.TrimSuffix(content, " then")
	comp := ""
	leftExpr := ""
	rightExpr := ""
	comparators := []string{"is higher than", "is lower than", "is not", "is"}
	for _, c := range comparators {
		idx := strings.Index(content, c)
		if idx != -1 {
			comp = c
			leftExpr = strings.TrimSpace(content[:idx])
			rightExpr = strings.TrimSpace(content[idx+len(c):])
			break
		}
	}
	if comp == "" {
		fmt.Println("[ERROR] Invalid if condition! Use: if <x> is higher than <y> then")
		return startIndex + 1
	}
	leftVal := evaluateMathWithVars(leftExpr, v)
	rightVal := evaluateMathWithVars(rightExpr, v)
	lNum, lErr := strconv.Atoi(leftVal)
	rNum, rErr := strconv.Atoi(rightVal)
	isTrue := false
	if lErr == nil && rErr == nil {
		switch comp {
		case "is higher than":
			isTrue = lNum > rNum
		case "is lower than":
			isTrue = lNum < rNum
		case "is":
			isTrue = lNum == rNum
		case "is not":
			isTrue = lNum != rNum
		}
	} else {
		switch comp {
		case "is":
			isTrue = leftVal == rightVal
		case "is not":
			isTrue = leftVal != rightVal
		}
	}
	i := startIndex + 1
	inElse := false
	for i < len(lines) {
		currentLine := strings.TrimSpace(lines[i])
		lowerCurrent := strings.ToLower(currentLine)
		if lowerCurrent == "else then" {
			if isTrue {
				i = findEndIf(lines, i)
				return i
			} else {
				inElse = true
				i++
				continue
			}
		}
		if lowerCurrent == "end if" {
			return i + 1
		}
		if (isTrue && !inElse) || (!isTrue && inElse) {
			if currentLine != "" && !strings.HasPrefix(currentLine, "#") {
				if strings.HasPrefix(lowerCurrent, "let ") {
					executeDefWithVars(currentLine, false, v)
				} else if strings.HasPrefix(lowerCurrent, "print ") {
					executePrintWithVars(strings.TrimPrefix(currentLine, "print "), false, v)
				} else if strings.HasPrefix(lowerCurrent, "repeat ") {
					i = handleRepeatBlockWithVars(lines, i, v)
					continue
				} else if strings.HasPrefix(lowerCurrent, "if ") {
					i = handleIfBlockWithVars(lines, i, v)
					continue
				} else if strings.HasPrefix(lowerCurrent, "answer ") {
					executeAnswerWithVars(currentLine, false, v)
				}
			}
		}
		i++
	}
	return i
}

func findEndIf(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.ToLower(strings.TrimSpace(lines[i])) == "end if" {
			return i + 1
		}
	}
	return len(lines)
}

func handleRepeatBlock(lines []string, startIndex int) int {
	return handleRepeatBlockWithVars(lines, startIndex, vars)
}

func handleRepeatBlockWithVars(lines []string, startIndex int, v map[string]string) int {
	line := strings.TrimSpace(lines[startIndex])
	content := strings.TrimPrefix(strings.ToLower(line), "repeat ")
	timesIdx := strings.Index(content, " times ")
	if timesIdx == -1 {
		fmt.Println("[ERROR] Invalid loop syntax! Use: repeat <n> times $var")
		return startIndex + 1
	}
	countExpr := strings.TrimSpace(content[:timesIdx])
	varName := strings.TrimSpace(content[timesIdx+len(" times "):])
	if !strings.HasPrefix(varName, "$") {
		fmt.Printf("[ERROR] Loop variable must start with '$'. Did you mean '$%s'?\n", varName)
		return startIndex + 1
	}
	countStr := evaluateMathWithVars(countExpr, v)
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 0 {
		fmt.Println("[ERROR] Invalid loop count!")
		return startIndex + 1
	}
	endIndex := findEndRepeat(lines, startIndex+1)
	if endIndex == -1 {
		fmt.Println("[ERROR] Missing 'end repeat'!")
		return len(lines)
	}
	key := varName[1:]
	for iteration := 1; iteration <= count; iteration++ {
		v[key] = strconv.Itoa(iteration)
		for j := startIndex + 1; j < endIndex; j++ {
			currentLine := strings.TrimSpace(lines[j])
			lowerCurrent := strings.ToLower(currentLine)
			if currentLine == "" || strings.HasPrefix(currentLine, "#") {
				continue
			}
			if strings.HasPrefix(lowerCurrent, "let ") {
				executeDefWithVars(currentLine, false, v)
			} else if strings.HasPrefix(lowerCurrent, "print ") {
				executePrintWithVars(strings.TrimPrefix(currentLine, "print "), false, v)
			} else if strings.HasPrefix(lowerCurrent, "if ") {
				j = handleIfBlockWithVars(lines, j, v)
				j--
			} else if strings.HasPrefix(lowerCurrent, "repeat ") {
				j = handleRepeatBlockWithVars(lines, j, v)
				j--
			} else if strings.HasPrefix(lowerCurrent, "answer ") {
				executeAnswerWithVars(currentLine, false, v)
			}
		}
	}
	delete(v, key)
	return endIndex + 1
}

func findEndRepeat(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.ToLower(strings.TrimSpace(lines[i])) == "end repeat" {
			return i
		}
	}
	return -1
}

// ✅ EXEC COMMAND ROUTER
func handleExecCommand(cmd string, interactive bool) {
	if strings.HasPrefix(strings.ToLower(cmd), "let ") {
		executeDef(cmd, interactive)
	} else if strings.HasPrefix(strings.ToLower(cmd), "print ") {
		executePrint(strings.TrimPrefix(cmd, "print "), interactive)
	} else if strings.HasPrefix(strings.ToLower(cmd), "answer ") {
		executeAnswer(cmd, interactive)
	} else {
		fmt.Println("[ERROR] Unknown command. Try 'lowell exec let ...' or 'lowell exec print ...'")
	}
}