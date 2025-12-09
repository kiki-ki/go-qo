# theme.tape

Output "theme_dark.gif"

Set Theme "Dracula"

Set Shell "zsh"
Set FontSize 18
Set Width 800
Set Height 400
Set Padding 20
Set TypingSpeed 50ms

Type "cat data/users.json | ../../bin/qo"
Enter
Sleep 2s

Ctrl+U
Type "SELECT name, city"
Sleep 1.5s
Type " FROM tmp WHERE age >= 30"
Sleep 1.5s
