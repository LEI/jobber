package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/dshearer/jobber/ipc"
)

func doPauseCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(PauseCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(PauseCmdStr, "[JOBS...]", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	// get jobs
	var jobs []string = flagSet.Args()

	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// send command
	var resp ipc.PauseCmdResp
	err = CallDaemon(
		"IpcService.Pause",
		ipc.PauseCmd{Jobs: jobs},
		&resp,
		usr,
		true,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	fmt.Printf("Paused %v jobs.\n", resp.NumPaused)
	return 0
}
