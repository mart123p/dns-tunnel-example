// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package main

//No need to hide the console on unix systems
func hideConsole() {
	return
}
