package config

import (
	"fmt"
	"os"
)

func ErrorInit(msg string) {
	fmt.Println("========================================")
	fmt.Println("❌ 配置初始化失败")
	fmt.Println("========================================")
	fmt.Printf("错误详情: %s\n", msg)
	fmt.Println()
	fmt.Println("💡 解决方案:")
	fmt.Println("1. 检查 .env 文件是否配置正确")
	fmt.Println("2. 检查 config.json 文件是否配置正确")
	fmt.Println()
	fmt.Println("按任意键退出程序...")

	// 等待用户输入
	var input string
	fmt.Scanln(&input)

	// 优雅退出
	os.Exit(1)
}
