Output assets/kubectl-peek-demo.gif

Set Shell "zsh"
Set FontSize 22
Set Width 1200
Set Height 700
Set TypingSpeed 60ms
Set Theme "Dracula"

# First run: built-in Secret discovery
Type "kubectl-peek secret"
Enter
Sleep 500ms
Right
Sleep 500ms
Right
Sleep 500ms

Type "/"
Sleep 500ms
Type "apm-server-apm-config"
Enter
Sleep 100ms

ScrollUp@250ms 40
Sleep 1s

# Exit kubectl-peek
Ctrl+C
Sleep 1s

# Show the custom rule file
Ctrl+L
Sleep 500ms
Type "cat $HOME/.config/kubectl-peek/rules.yaml"
Enter
Sleep 2s

ScrollUp@250ms 30
Sleep 2s

# Configure the rule file through the environment variable
Type "export KUBECTL_PEEK_RULE_FILE=$HOME/.config/kubectl-peek/rules.yaml"
Enter
Sleep 1s

# Second run: built-in and custom rule discovery
Type "clear"
Enter
Type "kubectl-peek secret -n airflow"
Enter
Right
Right

Type "/"
Type "airflow-metadata-db"
Enter
Sleep 100ms

ScrollUp@250ms 35
Sleep 1s
