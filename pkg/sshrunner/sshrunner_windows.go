//go:build windows
// +build windows

package sshrunner

import "golang.org/x/crypto/ssh"

func watchWindowSize(fd int, session *ssh.Session) {
	// Not implemented on Windows
}
