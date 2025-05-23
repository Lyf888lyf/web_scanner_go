package server

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser 尝试在当前操作系统上打开默认浏览器访问指定 URL
func OpenBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		fmt.Println("Error: unsupported platform for auto browser open")
		return
	}

	// 不阻塞主线程，处理执行错误
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error opening browser: %v\n", err)
	}
}
