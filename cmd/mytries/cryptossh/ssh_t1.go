package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// PasswordRecord 存储密码记录
type PasswordRecord struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"` // 实际应用中建议加密存储
}

// PasswordStore 密码存储管理
type PasswordStore struct {
	Records  []PasswordRecord `json:"records"`
	filePath string
}

// SSHConfig SSH 连接配置
type SSHConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

var (
	// 颜色输出
	cyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 解析命令参数
	config, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("错误:"), err)
		os.Exit(1)
	}

	// 初始化密码存储
	store, err := newPasswordStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("密码存储初始化失败:"), err)
		os.Exit(1)
	}

	// 查找已保存的密码
	if saved := store.find(config.Host, config.Port, config.Username); saved != nil {
		fmt.Printf("%s 发现已保存的密码记录 [%s@%s:%s]\n",
			cyan("→"), config.Username, config.Host, config.Port)
		fmt.Print("  使用已保存密码? [Y/n/s(删除记录)]: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "n", "no":
			// 继续请求新密码
		case "s", "delete":
			store.remove(config.Host, config.Port, config.Username)
			fmt.Println(yellow("  已删除密码记录"))
		default:
			config.Password = saved.Password
			fmt.Println(green("  使用已保存密码连接..."))
		}
	}

	// 如果没有密码，交互式请求
	if config.Password == "" {
		password, shouldSave, err := promptPassword(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", red("密码输入失败:"), err)
			os.Exit(1)
		}
		config.Password = password

		// 保存密码
		if shouldSave {
			store.add(PasswordRecord{
				Host:     config.Host,
				Port:     config.Port,
				Username: config.Username,
				Password: config.Password,
			})
			fmt.Println(green("✓ 密码已保存"))
		}
	}

	// 执行 SSH 连接
	if err := runSSH(config); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s %v\n", red("连接失败:"), err)
		os.Exit(1)
	}
}

// parseArgs 解析命令行参数，支持 ssh user@host -p port 格式
func parseArgs(args []string) (*SSHConfig, error) {
	config := &SSHConfig{
		Port: "22", // 默认端口
	}

	// 解析 user@host 格式
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		target := args[0]
		if strings.Contains(target, "@") {
			parts := strings.SplitN(target, "@", 2)
			config.Username = parts[0]
			config.Host = parts[1]
		} else {
			// 默认使用当前用户名
			currentUser, err := user.Current()
			if err != nil {
				return nil, fmt.Errorf("无法获取当前用户: %v", err)
			}
			config.Username = currentUser.Username
			config.Host = target
		}
	}

	// 解析 -p 端口参数
	for i := 1; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			config.Port = args[i+1]
			i++
		}
	}

	if config.Host == "" {
		return nil, fmt.Errorf("未指定目标主机")
	}

	return config, nil
}

// promptPassword 交互式密码输入，支持保存选项
func promptPassword(config *SSHConfig) (string, bool, error) {
	fmt.Printf("\n%s 连接到 %s@%s:%s\n", cyan("SSH"), config.Username, config.Host, config.Port)
	fmt.Println(strings.Repeat("─", 50))

	// 使用 readline 提供更好的输入体验
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          yellow("密码: "),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		EnableMask:      true, // 密码掩码
		MaskRune:        '*',
	})
	if err != nil {
		// 回退到标准终端输入
		return fallbackPasswordPrompt()
	}
	defer rl.Close()

	// 读取密码
	line, err := rl.Readline()
	if err != nil {
		return "", false, err
	}
	password := strings.TrimSpace(line)

	if password == "" {
		return "", false, fmt.Errorf("密码不能为空")
	}

	// 询问是否保存密码
	fmt.Print("保存密码以便下次自动填充? [y/N]: ")
	rl2, _ := readline.New("")
	saveChoice, _ := rl2.Readline()
	shouldSave := strings.ToLower(strings.TrimSpace(saveChoice)) == "y"

	return password, shouldSave, nil
}

// fallbackPasswordPrompt 标准终端密码输入（备用方案）
func fallbackPasswordPrompt() (string, bool, error) {
	fmt.Print(yellow("密码: "))
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", false, err
	}
	fmt.Println()

	password := string(bytePassword)
	if password == "" {
		return "", false, fmt.Errorf("密码不能为空")
	}

	fmt.Print("保存密码以便下次自动填充? [y/N]: ")
	var saveChoice string
	fmt.Scanln(&saveChoice)
	shouldSave := strings.ToLower(strings.TrimSpace(saveChoice)) == "y"

	return password, shouldSave, nil
}

// runSSH 执行 SSH 连接并处理交互
func runSSH(config *SSHConfig) error {
	// SSH 客户端配置
	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应使用 known_hosts 验证
		Timeout:         0,                           // 无超时，保持连接
	}

	// 建立连接
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	fmt.Printf("\n%s 正在连接 %s...\n", cyan("→"), addr)

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer client.Close()

	fmt.Printf("%s 认证成功！\n\n", green("✓"))

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()

	// 设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 开启回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度
		ssh.TTY_OP_OSPEED: 14400, // 输出速度
	}

	// 获取当前终端大小
	width, height, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	// 请求伪终端
	if err := session.RequestPty("xterm-256color", height, width, modes); err != nil {
		return fmt.Errorf("请求伪终端失败: %v", err)
	}

	// 设置标准输入输出
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	// 启动 shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("启动 shell 失败: %v", err)
	}

	// 处理窗口大小变化
	go handleWindowResize(session)

	// 等待会话结束
	if err := session.Wait(); err != nil {
		if err != io.EOF {
			return fmt.Errorf("会话异常: %v", err)
		}
	}

	return nil
}

// handleWindowResize 处理终端窗口大小变化
func handleWindowResize(session *ssh.Session) {
	// 简化实现，实际可监听 SIGWINCH 信号
	// 这里仅作为占位，完整实现需要信号处理
}

// newPasswordStore 初始化密码存储
func newPasswordStore() (*PasswordStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	store := &PasswordStore{
		filePath: filepath.Join(homeDir, ".ssh", "gossh_records.json"),
	}

	// 确保 .ssh 目录存在
	sshDir := filepath.Dir(store.filePath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, err
	}

	// 加载已有记录
	if _, err := os.Stat(store.filePath); err == nil {
		data, err := os.ReadFile(store.filePath)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(data, &store.Records)
	}

	return store, nil
}

// find 查找密码记录
func (s *PasswordStore) find(host, port, username string) *PasswordRecord {
	for _, r := range s.Records {
		if r.Host == host && r.Port == port && r.Username == username {
			return &r
		}
	}
	return nil
}

// add 添加或更新密码记录
func (s *PasswordStore) add(record PasswordRecord) {
	// 删除旧记录
	s.remove(record.Host, record.Port, record.Username)
	// 添加新记录
	s.Records = append(s.Records, record)
	s.save()
}

// remove 删除密码记录
func (s *PasswordStore) remove(host, port, username string) {
	newRecords := make([]PasswordRecord, 0, len(s.Records))
	for _, r := range s.Records {
		if !(r.Host == host && r.Port == port && r.Username == username) {
			newRecords = append(newRecords, r)
		}
	}
	s.Records = newRecords
	s.save()
}

// save 保存到文件
func (s *PasswordStore) save() error {
	data, err := json.MarshalIndent(s.Records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println(cyan("GoSSH - 智能 SSH 客户端工具"))
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  gossh <user@host> [-p port]")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  gossh root@192.168.1.1")
	fmt.Println("  gossh admin@server.com -p 2222")
	fmt.Println("  gossh 192.168.1.1          # 使用当前用户名")
	fmt.Println()
	fmt.Println("功能:")
	fmt.Println("  • 自动检测已保存的密码")
	fmt.Println("  • 交互式密码输入（隐藏显示）")
	fmt.Println("  • 可选保存密码以便下次自动填充")
	fmt.Println("  • 完整的终端交互支持")
	fmt.Println()
	fmt.Println(yellow("注意: 密码以明文形式存储在 ~/.ssh/gossh_records.json"))
	fmt.Println(yellow("      建议设置文件权限为 600，并在共享环境中谨慎使用"))
}
