Output assets/kubectl-peek-shell-workflows.gif

Set Shell "zsh"
Set FontSize 18
Set Width 1800
Set Height 800
Set TypingSpeed 60ms
Set Theme "Dracula"

Require tmux
Require kubectl-peek

Hide
Type "tmux kill-session -t kubectl-peek-demo 2>/dev/null || true"
Enter
Sleep 300ms

Type "tmux new-session -d -s kubectl-peek-demo"
Enter
Sleep 500ms

Type "tmux split-window -h -t kubectl-peek-demo:0"
Enter
Sleep 500ms

Type "tmux select-pane -t kubectl-peek-demo:0.0"
Enter
Sleep 500ms

Type "tmux attach-session -t kubectl-peek-demo"
Enter
Sleep 1s

Type "clear"
Enter
Show

Type "kubectl-peek shell"
Enter
Sleep 800ms

# Select the first available context.
Enter
Sleep 800ms

# Select the first available namespace.
Enter
Sleep 1500ms

Type "kubectl get pods  | grep Running"
Enter
Sleep 100ms

# Move to the right pane
Hide
Type "tmux select-pane -R"
Enter
Show

Type "kubectl-peek shell"
Enter
Sleep 800ms


# Select the first available context.
Enter
Sleep 800ms

Type "/"
Type "argocd"
Enter
Sleep 500ms

Type "kubectl get pods"
Enter
Sleep 100ms

Enter
Sleep 1200ms
