# [GENERAL]
Lowell Stable 1.0 is a human-first programming language.
- Variables start with $ (e.g., $User_Name).
- Math happens in < > (e.g., <$a + $b>).
- Text gluing happens in ( ) (e.g., ($name)+(" is cool")).
- Comments start with #.
- Use 'lowell repl' for interactive mode. Type 'exit' to quit.
- Dual-context prompt shows CWD + Lowell mode (e.g., C:\Path lowell>).

# [MATH]
Use angle brackets for computation: <$a + $b>.
- Spaces are ignored: <$a+$b> works too!
- Use $ to access variables inside math.
- Lowell auto-converts text numbers to real numbers in < >.
- Division by zero returns [ERROR], does not crash.

# [VARS]
Define variables using "let...be":
- RECOMMENDED: let $score be <100> (Safe for numbers!)
- ALLOWED: let $score be 100 (Works, but < > prevents bugs)
- REQUIRED: let $name be "Wallace" (Quotes for text!)
- View all vars: type 'vars' in REPL (shows numbered list).
- Delete vars: del $x | del $temp* | del 1-3 (index range).

# [LOGIC]
Make decisions with English words:
- if <$score> is higher than 80 then
    print "A+"
else then
    print "B"
end if
- Comparators: is higher than, is lower than, is, is not.
- Multi-line blocks ONLY work in .low files (not REPL).

# [LOOPS]
Repeat actions easily:
- repeat 5 times $i
    print $i
end repeat
- Loop var MUST start with $. Count can be math: repeat <$x+2> times $j.
- Loop var is deleted after loop ends.

# [INPUT]
Talk to your program:
- answer $name be "Who are you? " (Text input)
- answer $age be <How old are you?> (Number input)
- In isolated run (--iso), 'answer' is skipped with warning.

# [NAVIGATION]
Move through directories with plain English!
- navigate <path> → Change to specified directory.
- nvt <path>     → Shortcut alias for 'navigate' (faster typing).
- navigate       → Go directly to home directory.
- Auto-suggests similar folder names on typo!
- Prompt updates instantly to show new location.

# [EDIT]
Seamless file creation & editing workflow!
- edit <file.low> → Opens existing OR creates blank new one.
- et <file.low>   → Shortcut alias for 'edit' (faster typing).
- Auto-appends '.low' if extension missing.
- NON-BLOCKING: REPL stays alive while you edit!
- After saving, use 'run <file>' to test immediately.
- Fallback editor: VS Code > Notepad++ > Notepad (Windows).

# [RUN]
Execute .low files with precision state control!
- run <file.low>           → Executes in CURRENT session (vars persist).
- run --iso / riso <file>  → Executes in ISOLATED sandbox (no side effects).
- Both modes support relative paths (based on current prompt directory).
- File execution NEVER shows warnings or prompts (silent batch mode).

# [CLEAR]
Surgical state management for REPL sessions!
- clear                      → Warns + asks for confirmation (SAFE).
- clear --confirm / --cm     → Nukes ALL state instantly (FAST).
- clear --keepvars / --kvs   → Clears screen/history ONLY (PRESERVES VARS).
- clear --onlyvars / --ovs   → Wipes VARIABLES, keeps aliases (TARGETED).
- clear --shortcuts / --scs  → Wipes ALIASES, keeps variables (PRECISION).
- In .low files, ALL clear commands execute SILENTLY (no prompts).

# [HISTORY]
Searchable command memory for REPL sessions!
- history                    → Shows ALL commands.
- history N                  → Shows last N commands (e.g., history 10).
- history --kwd / --keyword "term" → Searches commands containing term.
- history --dwd / --download file.txt → Exports full history to file.
- History is REPL-ONLY. .low files NEVER track history.
- Search is case-insensitive.

# [SHORTCUTS]
Natural language aliases for personalized speed!
- shortcut "<command>" into "<alias>" → Creates persistent alias.
  Example: shortcut "history --kwd" into "hk"
- shortcutlist / sclist               → Shows all active shortcuts.
- shortcut delete <alias>             → Removes specific shortcut.
- Shortcuts ONLY work in REPL. .low files ALWAYS use full commands.
- Aliases are session-specific (lost on exit unless saved).

# [ERRORS]
Lowell gives smart errors:
- Missing $? → "Did you mean '$wal'?"
- Type mismatch? → "Use () to glue text, not < >!"
- Undefined var? → "Use 'let $x be <val>' to define it!"
- Invalid location? → "💡 Did you mean: [Downloads]?"
- Invalid syntax? → Suggests correct format with example.