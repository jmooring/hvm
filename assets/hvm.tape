# Create the terminal demonstration:
#
#   go install github.com/charmbracelet/vhs@latest
#   sudo snap install ttyd --classic
#   vhs hvm.tape
#
# Then compress with gif2webp:
#
#   gif2webp -mixed -m 6 -q 50 -kmin 0 -kmax 0 -min_size hvm.gif -o hvm.webp
#
# -----------------------------------------------------------------------------

# Output
Output hvm.gif

# Settings
Set FontSize 24
Set Framerate 24
Set Height 640
Set Padding 36
Set Width 1300
Set TypingSpeed 75ms

# Set prompt and clean cache
Hide
Type "hvm clean"
Enter
Type "y"
Enter
Type "source ~/.bashrc"
Enter
Type "PS1='\033[33m$\033[0m '"
Enter
Ctrl+L
Show

# Commands
Type "hvm use latest"
Sleep 1
Enter
Wait+Line@10s /\$/

Type "hugo version"
Sleep 1
Enter
Sleep 3

Type "hvm use"
Sleep 1
Enter
Sleep 3

Type "7"
Sleep 1
Enter
Wait+Line@10s /\$/

Type "hugo version"
Sleep 1
Enter
Sleep 3

Type "hvm status"
Sleep 1
Enter
Sleep 5
