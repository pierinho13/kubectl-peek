Output assets/kubectl-peek-demo.gif

Set Shell "zsh"
Set FontSize 22
Set Width 1200
Set Height 700
Set TypingSpeed 60ms
Set Theme "Dracula"

# First run: built-in Secret discovery
Type "kubectl-peek"
Enter
Sleep 2s

Right
Sleep 500ms
Right
Sleep 500ms

Type "/"
Sleep 500ms
Type "apm-server-apm-config"
Sleep 2s

Enter
Sleep 3s

ScrollUp@250ms 12
Sleep 1s
ScrollDown@250ms 12
Sleep 3s

# Exit kubectl-peek
Ctrl+C
Sleep 1s

# Show the custom rule file
Ctrl+L
Sleep 500ms
Type "cat $HOME/.config/kubectl-peek/rules.yaml"
Enter
Sleep 2s

ScrollUp@250ms 10
Sleep 1s
ScrollDown@250ms 10
Sleep 2s

# Configure the rule file through the environment variable
Type "export KUBECTL_PEEK_RULE_FILE=$HOME/.config/kubectl-peek/rules.yaml"
Enter
Sleep 1s

# Second run: built-in and custom rule discovery
Ctrl+L
Sleep 500ms
Type "kubectl-peek -n airflow"
Enter
Sleep 2s

Right
Sleep 500ms
Right
Sleep 500ms

Type "/"
Sleep 500ms
Type "airflow-metadata-db"
Sleep 2s

Enter
Sleep 3s

ScrollUp@250ms 12
Sleep 1s
ScrollDown@250ms 12
Sleep 3s
