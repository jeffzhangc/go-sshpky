package cmd

import (
	"fmt"
	"io"
	"os"
	"sshpky/pkg/sshrunner"
	"strings"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <host_alias> -- '<command>'",
	Short: "Execute a command on a remote host",
	Long: `Execute a command on a remote host using managed SSH keys.
The command leverages the existing host alias configuration.

Examples:
  # Execute a simple command
  sshpky exec my-server -- 'ls -la'

  # Execute a local script on the remote server
  sshpky exec my-server -f ./script.sh`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		execFile, _ := cmd.Flags().GetString("file")
		hostAlias := args[0]
		var scriptContent io.Reader
		var err error

		var commandToRun string
		dashPos := cmd.ArgsLenAtDash()

		if dashPos != -1 {
			commandParts := args[dashPos:]
			var b strings.Builder
			for i, p := range commandParts {
				if i > 0 {
					b.WriteString(" ")
				}
				// Basic shell quoting. This is not foolproof but handles many cases.
				if strings.ContainsAny(p, " '\"\\`$*?[]{}()<>|&;") {
					b.WriteString("'")
					b.WriteString(strings.ReplaceAll(p, "'", `'\''`))
					b.WriteString("'")
				} else {
					b.WriteString(p)
				}
			}
			commandToRun = b.String()
		} else if len(args) > 1 {
			// Handle case where user does `sshpky exec host command` without `--`.
			commandParts := args[1:]
			commandToRun = strings.Join(commandParts, " ")
		}

		if execFile != "" {
			if commandToRun != "" {
				fmt.Fprintln(os.Stderr, "Error: -f/--file and an inline command cannot be used at the same time.")
				osExit(1)
			}
			file, err := os.Open(execFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening script file: %v\n", err)
				osExit(1)
			}
			defer file.Close()
			scriptContent = file
		}

		if scriptContent == nil && commandToRun == "" {
			fmt.Fprintln(os.Stderr, "Error: you must provide either a command to run or a script file with -f/--file.")
			osExit(1)
		}

		exitCode, err := sshrunner.RunCommand(hostAlias, commandToRun, scriptContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		osExit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringP("file", "f", "", "local script file to execute on the remote host")
}
