//go:build !windows
// +build !windows

package sshrunner

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func watchWindowSize(fd int, session *ssh.Session) {
	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	go func() {
		for {
			<-sigwinch
			w, h, _ := term.GetSize(fd)
			session.WindowChange(h, w)
		}
	}()
}
