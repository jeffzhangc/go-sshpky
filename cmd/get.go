package cmd

import (
	"encoding/json"
	"fmt"
	"sshpky/pkg/config"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	outputFormat string
	specificKey  string
)

var getCmd = &cobra.Command{
	Use:   "get [HOSTNAME]",
	Short: "Get SSH configuration for a specific host",
	Long:  `Retrieves and displays the SSH configuration for a given host.`, 
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: getValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		group, _ := cmd.Flags().GetString("group")

		manager := config.NewSSHConfigManager()
		if group == "" {
			cfg := config.GetConfig()
			group = cfg.Use
		}

		configs, err := manager.GetConfigsByGroup(group)
		if err != nil {
			fmt.Printf("Error getting configs for group '%s': %v\n", group, err)
			return
		}

		var sshConfig *config.SshConfigItem
		for _, c := range configs {
			if c.Host == host {
				sshConfig = c
				break
			}
		}

		if sshConfig == nil {
			fmt.Printf("No configuration found for host '%s' in group '%s'\n", host, group)
			return
		}

		if specificKey != "" {
			printSpecificKey(sshConfig, specificKey)
			return
		}

		switch strings.ToLower(outputFormat) {
		case "json":
			printJSON(sshConfig)
		case "yaml":
			printYAML(sshConfig)
		default:
			printDefault(sshConfig)
		}
	},
}

func getValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return connValidArgs(cmd, args, toComplete)
}

func printSpecificKey(sshConfig *config.SshConfigItem, key string) {
	switch strings.ToLower(key) {
	case "password":
		fmt.Println(sshConfig.GetPassword())
	case "mfasecret":
		fmt.Println(sshConfig.GetMfaSecret())
	case "host":
		fmt.Println(sshConfig.Host)
	case "hostname":
		fmt.Println(sshConfig.HostName)
	case "port":
		fmt.Println(sshConfig.Port)
	case "user":
		fmt.Println(sshConfig.User)
	case "identityfile":
		fmt.Println(sshConfig.IdentityFile)
	case "group":
		fmt.Println(sshConfig.Group)
	case "desc":
		fmt.Println(sshConfig.Desc)
	default:
		fmt.Printf("Unknown key: %s\n", key)
	}
}

func printJSON(sshConfig *config.SshConfigItem) {
	b, err := json.MarshalIndent(sshConfig, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting to JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

func printYAML(sshConfig *config.SshConfigItem) {
	b, err := yaml.Marshal(sshConfig)
	if err != nil {
		fmt.Printf("Error formatting to YAML: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

func printDefault(sshConfig *config.SshConfigItem) {
	fmt.Printf("Host: %s\n", sshConfig.Host)
	fmt.Printf("HostName: %s\n", sshConfig.HostName)
	fmt.Printf("Port: %d\n", sshConfig.Port)
	fmt.Printf("User: %s\n", sshConfig.User)
	fmt.Printf("IdentityFile: %s\n", sshConfig.IdentityFile)
	fmt.Printf("Group: %s\n", sshConfig.Group)
	fmt.Printf("Description: %s\n", sshConfig.Desc)
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&outputFormat, "output", "o", "default", "Output format (default, json, yaml)")
	getCmd.Flags().StringVarP(&specificKey, "key", "k", "", "Get a specific key value (e.g., password, mfasecret)")
	getCmd.Flags().StringP("group", "g", "", "group for this ssh")
	getCmd.RegisterFlagCompletionFunc("group", groupFlagValidArgs)
}
