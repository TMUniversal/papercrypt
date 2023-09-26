# VHS Tape: PaperCrypt Demo

Output examples/demo/demo.gif
Output examples/demo/demo.webm

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

# Type "echo 'My Secret Message' > secret.txt" Sleep 500ms Enter Sleep 50ms
# Type "papercrypt generate -i secret.txt -o secret.pdf" Sleep 1000ms Enter Sleep 350ms
Type "echo 'secret' | papercrypt generate -o secret.pdf" Sleep 1000ms Enter Sleep 350ms
Type "example" Sleep 1000ms Enter Sleep 50ms
Type "example" Sleep 750ms Enter Sleep 50ms

Sleep 3s

Type "# The PDF cotains a QR code," Sleep 250ms Enter Sleep 50ms
Type "# which we extract with pdftoppm" Sleep 500ms Enter Sleep 50ms
Type "# Now we'll use the QR code to decode the message" Sleep 500ms Enter Sleep 50ms
Type "pdftoppm -png -r 300 secret.pdf secret" Sleep 500ms Enter Sleep 1000ms
Type "papercrypt qr -i secret-1.png -o secret.pc.txt" Sleep 500ms Enter Sleep 50ms

Sleep 2s

Type "papercrypt decode -i secret.pc.txt -o secret.decrypted.txt" Sleep 500ms Enter Sleep 50ms
Type "example" Sleep 1000ms
Screenshot examples/demo/decode.png
Enter Sleep 50ms

Type "cat secret.decrypted.txt" Sleep 500ms Enter Sleep 50ms


Sleep 5s
# Cleanup
Hide
Backspace@0 100
Type "rm -f secret.txt secret.pdf secret.pc.txt secret.decrypted.txt secret-1.png secret-2.png" Sleep 500ms Enter Sleep 50ms

