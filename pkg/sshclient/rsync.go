package sshclient

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

// RsyncOptions 定义 rsync 操作的选项
type RsyncOptions struct {
	Archive bool // 归档模式，保留权限和时间
	Verbose bool // 详细输出
	Delete  bool // 删除目标目录中多余的文件
}

// SyncLocalToRemote 将本地文件或目录同步到远程服务器
func SyncLocalToRemote(sftpClient *sftp.Client, localPath, remotePath string, opts RsyncOptions) error {
	lstat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("cannot stat local path %s: %w", localPath, err)
	}

	// 如果远程路径存在且是目录，则上传到该目录下
	rstat, err := sftpClient.Stat(remotePath)
	if err == nil && rstat.IsDir() {
		remotePath = sftp.Join(remotePath, filepath.Base(localPath))
	}

	if !lstat.IsDir() {
		return uploadFile(sftpClient, localPath, remotePath, opts)
	}

	// 遍历本地目录并同步到远程
	err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(localPath, path)
		destPath := sftp.Join(remotePath, filepath.ToSlash(relPath))

		if info.IsDir() {
			// 确保远程目录存在
			if _, err := sftpClient.Stat(destPath); os.IsNotExist(err) {
				if opts.Verbose {
					fmt.Printf("creating directory %s\n", destPath)
				}
				if err := sftpClient.MkdirAll(destPath); err != nil {
					return err
				}
			}
			return nil
		}

		// 是文件，检查是否需要同步
		remoteInfo, err := sftpClient.Stat(destPath)
		if err == nil && !info.ModTime().After(remoteInfo.ModTime()) {
			return nil // 不是更新，跳过
		}
		return uploadFile(sftpClient, path, destPath, opts)
	})
	if err != nil {
		return err
	}

	// 处理 --delete 选项
	if opts.Delete {
		return deleteExtraRemoteFiles(sftpClient, localPath, remotePath, opts)
	}
	return nil
}

// uploadFile 上传单个文件
func uploadFile(sftpClient *sftp.Client, localFile, remoteFile string, opts RsyncOptions) error {
	if opts.Verbose {
		fmt.Printf("uploading %s to %s\n", localFile, remoteFile)
	}

	info, err := os.Stat(localFile)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remoteFile)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if opts.Archive {
		if err := dstFile.Chmod(info.Mode()); err != nil {
			return err
		}
		// dstFile.Chtimes(info.ModTime(), info.ModTime())
	}
	return nil
}

// deleteExtraRemoteFiles 删除远程目录中本地不存在的文件
func deleteExtraRemoteFiles(sftpClient *sftp.Client, localBase, remoteBase string, opts RsyncOptions) error {
	if opts.Verbose {
		fmt.Println("checking for extraneous files to delete on remote...")
	}
	var toDelete []string
	walker := sftpClient.Walk(remoteBase)
	for walker.Step() {
		if walker.Err() != nil {
			continue
		}
		remoteWalkPath := walker.Path()
		relPath, _ := filepath.Rel(filepath.FromSlash(remoteBase), filepath.FromSlash(remoteWalkPath))
		localCheckPath := filepath.Join(localBase, relPath)

		if _, err := os.Stat(localCheckPath); os.IsNotExist(err) {
			toDelete = append(toDelete, remoteWalkPath)
		}
	}
	// 先删除文件，再删除目录（逆序）
	for i := len(toDelete) - 1; i >= 0; i-- {
		path := toDelete[i]
		if opts.Verbose {
			fmt.Printf("deleting remote %s\n", path)
		}
		// 尝试作为文件或空目录删除，然后尝试作为目录删除（可能非空）
		err := sftpClient.Remove(path)
		if err != nil {
			sftpClient.RemoveDirectory(path)
		}
	}
	return nil
}

// SyncRemoteToLocal 将远程文件或目录同步到本地
func SyncRemoteToLocal(sftpClient *sftp.Client, remotePath, localPath string, opts RsyncOptions) error {
	rstat, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("cannot stat remote path %s: %w", remotePath, err)
	}

	// 如果本地路径存在且是目录，则下载到该目录下
	lstat, err := os.Stat(localPath)
	if err == nil && lstat.IsDir() {
		localPath = filepath.Join(localPath, filepath.Base(remotePath))
	}

	if !rstat.IsDir() {
		return downloadFile(sftpClient, remotePath, localPath, opts)
	}

	// 遍历远程目录并同步到本地
	walker := sftpClient.Walk(remotePath)
	for walker.Step() {
		if walker.Err() != nil {
			return walker.Err()
		}

		remoteWalkPath := walker.Path()
		relPath, _ := filepath.Rel(filepath.FromSlash(remotePath), filepath.FromSlash(remoteWalkPath))
		localDestPath := filepath.Join(localPath, relPath)

		// 获取远程文件信息
		remoteInfo, err := sftpClient.Stat(remoteWalkPath)
		if err != nil {
			return err
		}

		if remoteInfo.IsDir() {
			// 确保本地目录存在
			if _, err := os.Stat(localDestPath); os.IsNotExist(err) {
				if opts.Verbose {
					fmt.Printf("creating directory %s\n", localDestPath)
				}
				if err := os.MkdirAll(localDestPath, 0755); err != nil {
					return err
				}
			}
			continue
		}

		// 是文件，检查是否需要同步
		localInfo, err := os.Stat(localDestPath)
		if err == nil && !remoteInfo.ModTime().After(localInfo.ModTime()) {
			continue // 不是更新，跳过
		}

		if err := downloadFile(sftpClient, remoteWalkPath, localDestPath, opts); err != nil {
			return err
		}
	}

	// 处理 --delete 选项
	if opts.Delete {
		return deleteExtraLocalFiles(sftpClient, localPath, remotePath, opts)
	}
	return nil
}

// downloadFile 下载单个文件
func downloadFile(sftpClient *sftp.Client, remoteFile, localFile string, opts RsyncOptions) error {
	if opts.Verbose {
		fmt.Printf("downloading %s to %s\n", remoteFile, localFile)
	}

	remoteInfo, err := sftpClient.Stat(remoteFile)
	if err != nil {
		return err
	}

	// 确保 本地目录存在
	if err := os.MkdirAll(filepath.Dir(localFile), 0755); err != nil {
		return err
	}

	srcFile, err := sftpClient.Open(remoteFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if opts.Archive {
		if err := os.Chmod(localFile, remoteInfo.Mode()); err != nil {
			return err
		}
		// os.Chtimes(localFile, remoteInfo.ModTime(), remoteInfo.ModTime())
	}
	return nil
}

// deleteExtraLocalFiles 删除本地目录中远程不存在的文件
func deleteExtraLocalFiles(sftpClient *sftp.Client, localBase, remoteBase string, opts RsyncOptions) error {
	if opts.Verbose {
		fmt.Println("checking for extraneous files to delete on local...")
	}
	var toDelete []string

	err := filepath.Walk(localBase, func(localWalkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(localBase, localWalkPath)
		remoteCheckPath := filepath.ToSlash(filepath.Join(remoteBase, relPath))

		_, err = sftpClient.Stat(remoteCheckPath)
		if os.IsNotExist(err) || err != nil {
			toDelete = append(toDelete, localWalkPath)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 先删除文件，再删除目录（逆序）
	for i := len(toDelete) - 1; i >= 0; i-- {
		path := toDelete[i]
		if opts.Verbose {
			fmt.Printf("deleting local %s\n", path)
		}
		if err := os.Remove(path); err != nil {
			// 尝试作为目录删除
			os.RemoveAll(path)
		}
	}
	return nil
}
