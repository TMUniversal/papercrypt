# VHS Tape: PaperCrypt generate

Output examples/demo/generate.gif
Output examples/demo/generate.webm

Require echo
Require papercrypt

Set Theme "Monokai Pro"
Set Shell "bash"
Set FontSize 32
Set Width 1200
Set Height 600
Set Margin 16
Set MarginFill "#6B50FF"
Set BorderRadius 12
Set WindowBar "Colorful"

Set FontFamily "Fira Code"

Set Framerate 60

Type "# Let's encrypt a secret with papercrypt" Sleep 500ms Enter Sleep 50ms

Type "echo 'My Secret Message' > secret.txt" Sleep 500ms Enter Sleep 50ms
Type "papercrypt generate -i secret.txt -o secret.pdf" Sleep 1000ms Enter Sleep 350ms
Type "example" Sleep 1000ms Enter Sleep 50ms
Type "example"
Sleep 750ms
Screenshot examples/demo/generate.png
Enter
Sleep 50ms

Sleep 5s
# Cleanup
Hide
Backspace@0 100
Type "rm -f secret.txt secret.pdf" Enter

