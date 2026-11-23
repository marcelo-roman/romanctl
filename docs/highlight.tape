Output docs/demo.gif

Set Shell bash
Set FontFamily "JetBrains Mono"
Set FontSize 14
Set Width 1100
Set Height 640
Set Padding 18
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set CursorBlink false
Set Framerate 12

Hide
Type "export ROMANCTL_LANG=en ROMANCTL_PROJETOS=$HOME/romanctl-demo ROMANCTL_TEMPLATES=$PWD/templates PS1='\[\e[1;35m\]demo\[\e[0m\] \[\e[36m\]\W\[\e[0m\] $ ' && mkdir -p ~/romanctl-demo && cd ~/romanctl-demo && clear"
Enter
Sleep 1s
Show

Type "romanctl new shop --stack go --arch hexagonal,ddd --no-git"
Enter
Sleep 3s
Type "cd shop && romanctl add resource Product"
Enter
Sleep 3s
Type "clear && romanctl up -d"
Enter
Sleep 4s
Type `curl -s -X POST localhost:8090/products -d '{"name":"Pen"}'; echo`
Enter
Sleep 2s
Type "romanctl ui"
Enter
Sleep 3s
Tab
Sleep 2.5s
Tab
Sleep 2.5s
Tab
Sleep 2.5s
Type "q"
Sleep 1s
Type "clear && romanctl down && cd .. && romanctl rm shop -y"
Enter
Sleep 3s
