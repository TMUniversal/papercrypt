# VHS Tape: PaperCrypt key

Output examples/demo/key.gif
Output examples/demo/key.webm

Require echo
Require papercrypt

Set Theme "Monokai Pro"
Set Shell "bash"
Set FontSize 32
Set Width 1200
Set Height 500
Set Margin 20
Set MarginFill "#6B50FF"
Set BorderRadius 12
Set WindowBar "Colorful"

Set FontFamily "Fira Code"

Set Framerate 60

Type "papercrypt key -w 6" Sleep 1000ms Enter
Sleep 50ms
Screenshot examples/demo/key.png
# Sleep 3s
# Type "papercrypt k -w 6 -o key.txt" Sleep 1000ms Enter Sleep 50ms
# Type "cat key.txt && echo" Sleep 500ms Enter

Sleep 5s

Hide
Type "rm -f key.txt" Enter
