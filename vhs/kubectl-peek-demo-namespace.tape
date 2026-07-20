Output assets/kubectl-peek-multi-namespace-shell.gif

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

# Left pane: Traefik
Type "kubectl-peek namespace --shell"
Enter
Sleep 2s

Type "/"
Sleep 500ms
Type "traefik"
Sleep 1s

Down
Sleep 500ms
Enter
Sleep 2s

Type "kubectl get pods"
Enter
Sleep 3s

# Move to the right pane
Hide
Type "tmux select-pane -R"
Enter
Show

# Right pane: Airflow
Type "kubectl-peek namespace --shell"
Enter
Sleep 2s

Type "/"
Sleep 500ms
Type "airflow"
Sleep 1s

Enter
Sleep 2s

Type "kubectl get pods"
Enter
Sleep 4s