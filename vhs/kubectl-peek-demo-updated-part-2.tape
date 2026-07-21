Output assets/kubectl-peek-demo-part-2.gif

Set Shell "zsh"
Set FontSize 22
Set Width 1200
Set Height 700
Set TypingSpeed 60ms
Set Theme "Dracula"

# Show the custom rule file
Hide
Type "unset KUBECTL_PEEK_RULE_FILE"
Enter
Type "clear"
Enter
Sleep 500ms
Type "export KIND=ExternalSecret"
Enter
Type "clear"
Enter
Sleep 500ms
Show
# Configure the rule file through the environment variable
Type "export KUBECTL_PEEK_RULE_FILE=$HOME/.config/kubectl-peek/rules.yaml"
Enter
Sleep 1s
Type "cat $KUBECTL_PEEK_RULE_FILE | yq '.rules[] | select(.kind == strenv(KIND))'"
Enter
Sleep 5s

# Second run: built-in and custom rule discovery
Type "clear"
Enter
Type "kubectl-peek secret -n monitoring --show-usage --show-values=false"
Enter
Sleep 500ms
Enter
Sleep 10s
