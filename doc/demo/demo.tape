# VHS Demo for qo
# Run: vhs demo.tape

Output demo.gif

Set Shell "zsh"
Set FontSize 18
Set Width 900
Set Height 500
Set Padding 30
Set TypingSpeed 50ms
Set Theme "GruvboxDarkHard"
Set WindowBar Colorful
Set BorderRadius 12
Set Margin 20

Type "cat data/users.json"
Enter
Sleep 1s

Type "cat data/users.json | qo"
Sleep 1s
Enter
Sleep 2s

Ctrl+U
Type "SELECT name, city FROM tmp WHERE age >= 30"
Sleep 1.5s

Tab
Sleep 1.5s

Down
Sleep 500ms
Right
Sleep 500ms

Tab
Sleep 500ms
Enter
Sleep 2s

Type "qo data/users.json data/orders.json"
Sleep 1s
Enter
Sleep 2s

Ctrl+U
Type "SELECT u.name, o.product, o.amount FROM users u JOIN orders o ON u.id = o.user_id"
Sleep 1.5s

Tab
Sleep 1.5s

Down
Sleep 500ms
Down
Sleep 500ms
Right
Sleep 500ms

Tab
Sleep 500ms
Enter
Sleep 2s

Type "qo -o table -q 'SELECT name, age, city FROM users ORDER BY age DESC' data/users.json"
Sleep 1s
Enter
Sleep 3s
