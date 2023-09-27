<p align="center">
  <a href="https://goreportcard.com/report/github.com/fyne-io/terminal"><img src="https://goreportcard.com/badge/github.com/fyne-io/terminal" alt="Code Status" /></a>
  <a href="https://github.com/fyne-io/terminal/actions"><img src="https://github.com/fyne-io/terminal/workflows/Platform%20Tests/badge.svg" alt="Build Status" /></a>
  <a href='https://coveralls.io/github/fyne-io/terminal?branch=master'><img src='https://coveralls.io/repos/github/fyne-io/terminal/badge.svg?branch=master' alt='Coverage Status' /></a>
  <a href='http://gophers.slack.com/messages/fyne'><img src='https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=blue' alt='Join us on Slack' /></a>
</p>

# Fyne Terminal

A terminal emulator using the Fyne toolkit, supports Linux, macOS, Windows and BSD.

Running on Linux with a custom zsh theme.
<img alt="screenshot" src="img/linux.png" width="929" />

Running on macOS with a powerlevel10k zsh theme and classic style.
<img alt="screenshot" src="img/macos.png" width="912" />

Running on Windows with PowerShell running inside.
<img alt="screenshot" src="img/windows.png" width="900" />

# Installing on command line

Just use the go get command (you'll need a Go and C compiler installed first):

```
go get github.com/fyne-io/terminal/cmd/fyneterm
```

# Installing as an app

To get the app installed alongside your other applications (with metadata, icons etc),
use the `fyne` tool, as illustrated below:

```
$ go get fyne.io/fyne/v2/cmd/fyne
$ fyne get github.com/fyne-io/terminal/cmd/fyneterm
```

# TODO

There are lots of great things that could be added to this app.
Already planned is:

* Tabs
* Scroll-back
* Background and font/size customisation
* Split panels
* Windows support

# Library

You can also use this project as a library to create your own
terminal based applications, using the import path "github.com/fyne-io/terminal".

There are two modes, using the default shell (macOS and Linux) or connecting
to a remote shell.

## Local Shell

To load a terminal widget and launch the current shell (works on macOS and Linux)
use the `RunLocalShell` method after creating a `Terminal`, as follows:

```go
	// win is a fyne.Window created to hold the content
	t := terminal.New()
	w.SetContent(t)

	go func() {
		_ = t.RunLocalShell()
		a.Quit()
	}()
	w.ShowAndRun()
```

## Remote connection

For example to open a terminal to an SSH connection that you have created:

```go
	// session is an *ssh.Session from golang.org/x/crypto/ssh
	// win is a fyne.Window created to hold the content
	in, _ := session.StdinPipe()
	out, _ := session.StdoutPipe()

	go session.Run("$SHELL || bash")

	t := terminal.New()
	w.SetContent(t)

	go func() {
		_ = t.RunWithConnection(in, out)
		a.Quit()
	}()
	w.ShowAndRun()
```
