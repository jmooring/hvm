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

Output hvm.gif

# -----------------------------------------------------------------------------
# Settings
# -----------------------------------------------------------------------------

Set FontSize 24
Set Framerate 24
Set Height 640
Set Padding 36
Set Width 1300
Set TypingSpeed 75ms

# -----------------------------------------------------------------------------
# Reset hvm configuration and remove its cache
# -----------------------------------------------------------------------------

Hide
Type "hvm reset"
Enter
Type "y"
Enter
Type "source ~/.bashrc"
Enter
Type "PS1='\033[33m$\033[0m '"
Enter
Ctrl+L
Show

# -----------------------------------------------------------------------------
# Run hvm use
# -----------------------------------------------------------------------------

Type "hvm use"
Sleep 1
Enter
Wait+Line@10s /Select/ # wait for list to appear
Sleep 2

Type "1" # version
Sleep 1
Enter
Wait+Line@10s /Select/ # wait for list to appear
Sleep 2

Type "1" # edition
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

# -----------------------------------------------------------------------------
# Run hugo version
# -----------------------------------------------------------------------------

Type "hugo version"
Sleep 1
Enter
Sleep 2

# -----------------------------------------------------------------------------
# Run hvm use
# -----------------------------------------------------------------------------

Type "hvm use"
Sleep 1
Enter
Wait+Line@10s /Select/ # wait for list to appear
Sleep 2

Type "5" # version
Sleep 1
Enter
Wait+Line@10s /Select/ # wait for list to appear
Sleep 2

Type "2" # edition
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

# -----------------------------------------------------------------------------
# Run hugo version
# -----------------------------------------------------------------------------

Type "hugo version"
Sleep 1
Enter
Sleep 2

# -----------------------------------------------------------------------------
# Run hvm status
# -----------------------------------------------------------------------------

Type "hvm status"
Sleep 1
Enter
Sleep 2

# -----------------------------------------------------------------------------
# Wait a few more seconds before next loop
# -----------------------------------------------------------------------------

Sleep 4
