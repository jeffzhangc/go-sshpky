package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	// 颜色输出
	cyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
)

func main() {
	conf := SSHConfig{
		Host:     "example.com",
		Port:     "22",
		Username: "user",
	}

	a, b, e := promptPassword(&conf)
	fmt.Printf("密码: %s, 是否保存: %v, 错误: %v\n", a, b, e)

	xxx, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("读取密码时出错: %v\n", err)
	} else {
		fmt.Printf("读取到的密码长度: %d,%s\n", len(xxx), string(xxx))
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

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

// SSHConfig SSH 连接配置
type SSHConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

// fallbackPasswordPrompt 标准终端密码输入（备用方案）
func fallbackPasswordPrompt() (string, bool, error) {
	fmt.Print(yellow("密码: "))
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
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
