package cmd

import (
	"fmt"
	"os"

	"sshpky/pkg/config"
	sftpPkg "sshpky/pkg/sftp"
	"sshpky/pkg/sshrunner"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
)

var (
	rsyncArchive bool
	rsyncVerbose bool
	rsyncDelete  bool
)

// rsyncCmd represents the rsync command
var rsyncCmd = &cobra.Command{
	Use:   "rsync [-avz] [--delete] <source> <destination>",
	Short: "Sync files between local and remote using SFTP",
	Long: `Provides rsync-like file synchronization functionality using the SFTP protocol.
This avoids the need for the 'rsync' binary to be installed.

Remote paths are specified using the format: [user@]host_alias:path

Examples:
  # Upload local directory to remote server
  sshpky rsync -av ./local-dir my-server:/remote/app/

  # Upload with delete flag to remove extraneous files from destination
  sshpky rsync -av --delete ./local-dir my-server:/remote/app/
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfgManager := config.NewSSHConfigManager()

		// 解析源和目标路径
		sourceInfo := sftpPkg.ParseRemotePath(args[0])
		destInfo := sftpPkg.ParseRemotePath(args[1])

		// 验证源和目标不能同为远程或本地
		if sourceInfo.IsRemote == destInfo.IsRemote {
			if sourceInfo.IsRemote {
				fmt.Fprintln(os.Stderr, "Error: Source and destination cannot both be remote.")
			} else {
				fmt.Fprintln(os.Stderr, "Error: Source and destination cannot both be local.")
			}
			osExit(1)
			return
		}

		// 确定远程和本地路径信息
		var remote, local *sftpPkg.RemotePathInfo
		var isUpload bool
		if sourceInfo.IsRemote {
			remote, local = sourceInfo, destInfo
			isUpload = false
		} else {
			remote, local = destInfo, sourceInfo
			isUpload = true
		}

		// 查找 SSH 配置
		sshConfig, err := cfgManager.FindConfig(remote.HostAlias)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding SSH config for alias '%s': %v\n", remote.HostAlias, err)
			osExit(1)
			return
		}

		// 如果路径中指定了用户，覆盖配置中的用户
		if remote.User != "" {
			sshConfig.User = remote.User
		}

		// 建立 SSH 连接
		client, _, err := sshrunner.EstablishSSHClient(sshConfig.Host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error establishing SSH connection: %v\n", err)
			osExit(1)
			return
		}
		defer client.Close()

		// 创建 SFTP 客户端
		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating SFTP session: %v\n", err)
			osExit(1)
			return
		}
		defer sftpClient.Close()

		// 构建同步选项
		opts := sftpPkg.RsyncOptions{
			Archive: rsyncArchive,
			Verbose: rsyncVerbose,
			Delete:  rsyncDelete,
		}

		// 展开本地路径中的 ~
		localPath := sftpPkg.ExpandPath(local.Path)

		// 执行同步
		var syncErr error
		if isUpload {
			syncErr = sftpPkg.SyncLocalToRemote(sftpClient, localPath, remote.Path, opts)
		} else {
			syncErr = sftpPkg.SyncRemoteToLocal(sftpClient, remote.Path, localPath, opts)
		}

		if syncErr != nil {
			fmt.Fprintf(os.Stderr, "Synchronization failed: %v\n", syncErr)
			osExit(1)
			return
		}

		if opts.Verbose {
			fmt.Println("Synchronization completed successfully.")
		}
	},
}

func init() {
	rootCmd.AddCommand(rsyncCmd)
	rsyncCmd.Flags().BoolVarP(&rsyncArchive, "archive", "a", false, "archive mode; preserves permissions and times")
	rsyncCmd.Flags().BoolVarP(&rsyncVerbose, "verbose", "v", false, "increase verbosity")
	rsyncCmd.Flags().BoolVar(&rsyncDelete, "delete", false, "delete extraneous files from dest dirs")
}
