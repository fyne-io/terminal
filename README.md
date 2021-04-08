<p align="center">
  <a href="https://goreportcard.com/report/github.com/fyne-io/terminal"><img src="https://goreportcard.com/badge/github.com/fyne-io/terminal" alt="Code Status" /></a>
  <a href="https://travis-ci.org/fyne-io/terminal"><img src="https://travis-ci.org/fyne-io/terminal.svg" alt="Build Status" /></a>
  <a href='https://coveralls.io/github/fyne-io/terminal?branch=master'><img src='https://coveralls.io/repos/github/fyne-io/terminal/badge.svg?branch=master' alt='Coverage Status' /></a>
  <a href='http://gophers.slack.com/messages/fyne'><img src='https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=blue' alt='Join us on Slack' /></a>
</p>

# Fyne Terminal

A terminal emulator using the Fyne toolkit, supports Linux and Windows.

Running on Linux with a custom zsh theme.
<img alt="screenshot" src="img/linux.png" width="929" />

Running on macOS with a powerlevel10k zsh theme and classic style.
<img alt="screenshot" src="img/macos.png" width="912" />

# Installing

Just use the go get command (you'll need a Go and C compiler installed first):

```
$ go get github.com/fyne-io/terminal/cmd/fyneterm
```

# Library

You can also use this project as a library to create your own
terminal based applications, using the import path "github.com/fyne-io/terminal".

For example to open a terminal to an SSH connection that you have created:

```go
	// session is an *fynessh.Session from golang.org/x/crypto/fynessh
    // win is a fyne.Window created to hold the content
	in, _ := session.StdinPipe()
	out, _ := session.StdoutPipe()

	go session.Run("$SHELL || bash")

	t := terminal.NewTerminal()
	w.SetContent(t)

	go func() {
		_ = t.RunWithConnection(in, out)
		a.Quit()
	}()
	w.ShowAndRun()
```
