// +build !darwin

package main

import (
	"bufio"
	"github.com/dshearer/jobber/common"
	"os"
	"os/user"
	"strings"
)

/*
Get all users that have home dirs.
*/
func listUsers(prefs *Prefs) ([]*user.User, error) {
	users := make([]*user.User, 0)

	// open passwd
	f, err := os.Open("/etc/passwd")
	if err != nil {
		common.ErrLogger.Printf("Failed to open /etc/passwd: %v\n", err)
		return users, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// look up user
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) == 0 {
			continue
		}
		usr, err := user.Lookup(parts[0])
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
