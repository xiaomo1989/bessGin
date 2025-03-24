package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd 是 CLI 的根命令
var RootCmd = &cobra.Command{
	Use:   "artisan",
	Short: "Gin Artisan 命令行工具",
	Long:  "这个命令行工具类，用于管理和执行 Gin 相关任务",
}

// Execute 运行根命令
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println("执行 Artisan 命令时出错:", err)
		os.Exit(1)
	}
}
