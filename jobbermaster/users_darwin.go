// +build darwin

package main

import (
	"os/exec"
	"os/user"
	"strings"
)

/*
Get all users that have home dirs.
*/
func listUsers(prefs *Prefs) ([]*user.User, error) {
	users := make([]*user.User, 0)
	// dscl . list /Users | grep -v '^_'
	out, err := exec.Command("dscl", ".", "list", "/Users").Output()
	if err != nil {
		return users, err
	}
	for _, s := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(s, "_") {
			continue
		}
		// look up user
		usr, err := user.Lookup(s)
		if err != nil {
			continue
		}
		// check for reasons to exclude
		if !shouldRunForUser(usr, prefs) {
			continue
		}
		users = append(users, usr)
	}
	return users, nil
}
