# Create the terminal demonstration:
#
#   go install github.com/charmbracelet/vhs@latest
#   sudo snap install ttyd --classic
#   vhs hvm.tape

# Output
Output hvm.gif
#Output hvm.mp4
#Output hvm.webm

# Settings
Set FontSize 24
Set Framerate 24
Set Height 640
Set Padding 36
Set Width 1300

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
Type "hvm use"
Enter
Sleep 3

Type "1"
Enter
Sleep 5

Type "hugo version"
Enter
Sleep 3

Type "hvm use"
Enter
Sleep 3

Type "7"
Enter
Sleep 5

Type "hugo version"
Enter
Sleep 3

Type "hvm status"
Enter
Sleep 5
