package main

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var txt = "等待生成链接..."

func main() {

	RUN_UI() // 在子进程中执行UI函数

}

func UI() {

	var cmd *exec.Cmd
	// 确保在脚本关闭时终止后台运行的ngrok进程。这样可以避免ngrok进程在脚本关闭后继续运行。
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Println("Error killing ngrok process:", err)
		}
		fmt.Println("Ngrok process Stopped.")
	}()

	myApp := app.New()                          // 创建一个新的Fyne应用
	myWindow := myApp.NewWindow("Ngrok Config") // 创建一个新的窗口，标题为"Ngrok Config"
	myWindow.Resize(fyne.NewSize(300, 200))     // 设置窗口大小为300x200
	protocol := ""
	/* protocolEntry := widget.NewEntry()                        // 创建一个文本输入框用于输入协议
	protocolEntry.SetPlaceHolder("Enter protocol (http/tcp)") // 设置输入框的占位符文本 */
	protocolSelect := widget.NewSelect([]string{"http"}, func(value string) { // 创建一个选择框用于选择协议
		protocol = value // 更新选择的协议
	})
	protocolSelect.SetSelected("http")            // 设置默认选择为http
	portEntry := widget.NewEntry()                // 创建一个文本输入框用于末端口号
	portEntry.SetPlaceHolder("Enter port number") // 设置输入框的占位符文本

	port := ""
	// 在布局中添加新的标签组件
	urlLink := widget.NewHyperlink(txt, nil)

	submitButton := widget.NewButton("Start Ngrok", func() { // 创建一个按钮，点击后执行函数
		// 创建一个通道，用于传递协议和端口号
		protocol = protocolSelect.Selected                                         // 获取选择框的选择
		port = portEntry.Text                                                      // 获取端口号输入框的文本
		fmt.Printf("Starting ngrok with protocol: %s, port: %s\n", protocol, port) // 打印启动信息

		// 创建并执行命令对象
		cmd = exec.Command("ngrok", protocol, port)
		// 启动命令
		err := cmd.Start()
		// 检查命令是否成功启动
		if err != nil {
			fmt.Println("Error starting ngrok:", err)

			return
		}
		// 启动协程，执行爬虫并将结果写入通道
		RodGetUrl()

		parsedURL, _ := url.Parse(txt)

		urlLink.SetText(txt)
		urlLink.SetURL(parsedURL)

	})
	// 创建一个停止进程的按钮
	stopButton := widget.NewButton("Stop Ngrok", func() {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Println("Error killing ngrok process:", err)
		}
		fmt.Println("Ngrok process Stopped.")
		// 更新URL显示标签为"等待生成链接..."
		urlLink.SetText("等待生成链接...")
	})

	// 将数据写入到通道中

	/* 	ch <- protocol
	   	ch <- port */
	// ... existing code ...

	// 修改后的布局
	myWindow.SetContent(container.NewVBox(
		protocolSelect,
		portEntry,
		submitButton,

		urlLink,    // 添加URL显示标签
		stopButton, // 添加停止进程的按钮
	))

	myWindow.ShowAndRun() // 显示窗口并运行应用
}

func RUN_UI() {

	var cmd *exec.Cmd

	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Println("Error killing ngrok process:", err)
		}
		fmt.Println("Ngrok process Stopped.")
	}()

	myApp := app.New()
	myWindow := myApp.NewWindow("Ngrok Config")
	myWindow.Resize(fyne.NewSize(400, 350)) // 稍微放大窗口

	protocol := ""
	protocolSelect := widget.NewSelect([]string{"http", "tcp"}, func(value string) {
		protocol = value
	})
	protocolSelect.SetSelected("http")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("请输入端口号，例如: 8080")

	urlLink := widget.NewHyperlink(txt, nil)

	statusLabel := widget.NewLabel("当前状态：")
	// 创建一个启动按钮

	submitButton := widget.NewButton("🚀 启动 Ngrok", nil) // 先用 nil 占位
	stopButton := widget.NewButton("🛑 停止 Ngrok", nil)
	// 初始状态
	submitButton.Enable()
	stopButton.Disable()

	submitButton.OnTapped = func() {
		// 检查 ngrok 是否安装
		path, err := exec.LookPath("ngrok")
		if err != nil {
			statusLabel.SetText("❌ ngrok 未安装或未加入 PATH")
			fmt.Println("ngrok 未安装或未加入 PATH")
		} else {
			statusLabel.SetText("✅ ngrok 已安装")
			fmt.Println("ngrok 安装路径为:", path)
		}

		protocol = protocolSelect.Selected
		port := portEntry.Text

		// 端口验证
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1024 || portNum > 65535 {
			statusLabel.SetText("❌ 请输入有效的端口号（1024-65535）")
			return
		}

		fmt.Printf("Starting ngrok with protocol: %s, port: %s\n", protocol, port)

		// 检查是否已有ngrok进程
		cmdCheck := exec.Command("pgrep", "ngrok")
		output, err := cmdCheck.CombinedOutput()
		if err == nil && len(output) > 0 {
			statusLabel.SetText("❌ ngrok 已经在运行")
			// time.Sleep(3 * time.Second)
			// 杀死ngrok进程
			// 如果已有ngrok进程，尝试杀死它
			cmdKill := exec.Command("pkill", "ngrok") // 使用pkill命令杀死所有ngrok进程
			killOutput, killErr := cmdKill.CombinedOutput()
			if killErr != nil {
				statusLabel.SetText("❌ 无法终止已运行的ngrok进程")
				fmt.Println("杀死进程错误：", string(killOutput))
				return
			}

			// 杀死进程成功，提示并继续启动
			statusLabel.SetText("✅ 已终止现有 ngrok 进程，正在继续启动...")
			// time.Sleep(3 * time.Second)
		}

		// 使用命令执行文件的绝对路径
		cmd = exec.Command(path, protocol, port)
		err = cmd.Start()
		if err != nil {
			statusLabel.SetText("❌ Ngrok 启动失败：" + err.Error())
			fmt.Println("详细输出：" + err.Error())
			return
		}

		statusLabel.SetText("已启动，正在获取链接...")
		RodGetUrl()
		parsedURL, _ := url.Parse(txt)
		urlLink.SetText(txt)
		urlLink.SetURL(parsedURL)
		statusLabel.SetText("链接生成成功")
		// 状态更新
		submitButton.Disable()
		stopButton.Enable()
	}

	stopButton.OnTapped = func() {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Println("Error killing ngrok process:", err)
			statusLabel.SetText("停止失败")
			return
		}
		urlLink.SetText("等待生成链接...")
		statusLabel.SetText("Ngrok 已停止")
		// 状态更新
		submitButton.Enable()
		stopButton.Disable()
	}

	form := container.NewVBox(
		widget.NewLabelWithStyle("Ngrok 快速配置", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabel("选择协议："),
			protocolSelect,
			widget.NewLabel("输入端口号："),
			portEntry,
		),
		container.NewHBox(
			submitButton,
			stopButton,
		),
		statusLabel,
		urlLink,
	)

	card := container.NewVBox(
		container.NewCenter(form),
	)

	myWindow.SetContent(container.NewVBox(
		container.NewCenter(card),
	))

	myWindow.ShowAndRun()

}

func RodGetUrl() {

	// 禁用无头模式，启动浏览器
	url := launcher.New().Headless(true).MustLaunch()

	// 创建带有超时的上下文，设置为60秒
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 创建浏览器实例并连接
	browser := rod.New().ControlURL(url).Context(ctx).MustConnect()
	// 在程序结束时关闭浏览器
	defer browser.MustClose()

	page := browser.MustPage("http://localhost:4040/inspect/http")
	page.MustWaitLoad()
	// 获取元素
	element := page.MustElement("ul.tunnels li a")
	fmt.Println(element.MustText())
	// 将元素的文本输入到通道中
	txt = element.MustText()

}
