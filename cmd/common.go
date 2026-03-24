package cmd

import (
	"sshpky/pkg/config"
	"sshpky/pkg/utils"
	"strings"

	"github.com/spf13/cobra"
)

// ConnValidArgs provides command argument completion for host aliases.
func ConnValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	group, _ := cmd.Flags().GetString("group")
	manager := config.NewSSHConfigManager()
	var configs []*config.SshConfigItem
	var err error

	if group == "" {
		cfg := config.GetConfig()
		group = cfg.Use
	}

	if group != "" {
		configs, err = manager.GetConfigsByGroup(group)
	} else {
		configs, err = manager.ReadConfig()
	}

	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, config := range configs {
		suggestions = append(suggestions, config.Host)
	}

	var filtered []string
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, toComplete) {
			filtered = append(filtered, suggestion)
		}
	}

	if len(filtered) == 0 {
		filtered = suggestions
	}

	return filtered, cobra.ShellCompDirectiveNoFileComp
}

// GroupUseValidArgs provides command argument completion for group names.
func GroupUseValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg := config.GetConfig()
	var groups []string
	for _, g := range cfg.Groups {
		groups = append(groups, g.Name)
	}

	var filtered []string
	for _, group := range groups {
		if strings.HasPrefix(group, toComplete) {
			filtered = append(filtered, group)
		}
	}

	if len(filtered) == 0 {
		filtered = groups
	}

	return filtered, cobra.ShellCompDirectiveNoFileComp
}

// GroupFlagValidArgs provides flag completion for the 'group' flag.
func GroupFlagValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// For flags, args will be the list of arguments provided to the command,
	// so we don't check its length.
	return GroupUseValidArgs(cmd, []string{}, toComplete)
}

func RsyncValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// If completing the second argument...
	if len(args) == 1 {
		// and the first argument is not a remote path, then we can complete hosts for the second argument.
		sourceInfo := utils.ParseRemotePath(args[0])
		if sourceInfo.IsRemote {
			// first arg is remote, second must be local, enable file completion
			return nil, cobra.ShellCompDirectiveDefault
		}
	}

	// If completing more than the second argument, no completions.
	if len(args) >= 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Check if the input looks like a local path (starts with ./, /, or ~/)
	if strings.HasPrefix(toComplete, "./") || strings.HasPrefix(toComplete, "/") || strings.HasPrefix(toComplete, "~/") {
		// Enable file/directory completion for local paths
		return nil, cobra.ShellCompDirectiveDefault
	}

	// Completing first argument, or second argument when first is local.
	group, _ := cmd.Flags().GetString("group")
	manager := config.NewSSHConfigManager()
	var configs []*config.SshConfigItem
	var err error

	if group == "" {
		cfg := config.GetConfig()
		group = cfg.Use
	}

	if group != "" {
		configs, err = manager.GetConfigsByGroup(group)
	} else {
		configs, err = manager.ReadConfig()
	}

	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, config := range configs {
		suggestions = append(suggestions, config.Host+":")
	}

	// Standard filtering
	var filtered []string
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, toComplete) {
			filtered = append(filtered, suggestion)
		}
	}

	if len(filtered) == 0 {
		// if no direct match, maybe just show hosts without the colon
		for _, config := range configs {
			if strings.HasPrefix(config.Host, toComplete) {
				filtered = append(filtered, config.Host)
			}
		}
		if len(filtered) == 0 {
			filtered = suggestions // show with colon
		}
	}

	return filtered, cobra.ShellCompDirectiveNoFileComp
}
