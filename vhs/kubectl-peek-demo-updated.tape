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
Down
Down
Down
Down
Down
Down
Sleep 500ms
Enter

Type "kubectl-peek secret -n monitoring --show-usage"
Enter
Sleep 500ms
Type "/"
Sleep 500ms
Type "grafana"
Enter
Sleep 100ms

ScrollUp@250ms 100
Sleep 1s

# Exit kubectl-peek
Ctrl+C
Sleep 1s

# Show the custom rule file
Hide
Type "unset KUBECTL_PEEK_RULE_FILE"
Enter
Show
Type "clear"
Enter
Sleep 500ms
Type "more $HOME/.config/kubectl-peek/rules.yaml"
Enter
Enter
Enter
Type q 
Sleep 2s

# Configure the rule file through the environment variable
Type "export KUBECTL_PEEK_RULE_FILE=$HOME/.config/kubectl-peek/rules.yaml"
Enter
Sleep 1s

# Second run: built-in and custom rule discovery
Type "clear"
Enter
Type "kubectl-peek secret -n monitoring --show-usage"
Enter

Type "/"
Type "grafana"
Enter
Sleep 100ms

ScrollUp@250ms 100
Sleep 1s
