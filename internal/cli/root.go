package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/internal/handler"
	log "github.com/wxnacy/wgo/internal/logger"
	"github.com/wxnacy/wgo/internal/terminal"
)

var (
	logger    = log.GetLogger()
	startTime time.Time
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "wgo",
	Short:   "类 IPython 的 Golang 交互运行工具",
	Long:    ``,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		startTime = time.Now()
		// 初始化应用
		handler.Init()
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// handler.Destory()
		duration := time.Since(startTime)
		logger.Infof("命令执行耗时: %v\n", duration)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("wgo")
		// req := GetGlobalReq()
		// path := req.Path
		// if len(args) > 0 {
		// path = args[0]
		// }
		// if req.IsVerbose {
		// logger.SetLogLevel(logrus.DebugLevel)
		// }
		handleCmdErr(terminal.Run())
	},
}

var ErrQuit = errors.New("quit wgo")

func handleCmdErr(err error) {
	if err != nil {
		if err.Error() == "^D" ||
			err.Error() == "^C" ||
			err == ErrQuit {
			fmt.Println("GoodBye")
			os.Exit(0)
		}
		logger.Printf("Error: %v", err)
		logger.Errorf("Error: %v", err)
		os.Exit(0)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.PersistentFlags().BoolVarP(&globalReq.IsVerbose, "verbose", "v", false, "打印赘余信息")
	// rootCmd.PersistentFlags().StringVarP(&globalReq.Config, "config", "c", defaultConfig, "指定配置文件地址")

	// root 参数
	// rootCmd.PersistentFlags().StringVarP(&bdpanCommand.Path, "path", "p", "/", "直接查看文件")
	// rootCmd.PersistentFlags().StringVarP(&globalReq.Path, "path", "p", "/", "网盘文件地址")
	// rootCmd.PersistentFlags().IntVarP(&rootCommand.Limit, "limit", "l", 10, "查询数目")
	// 运行前全局命令
	cobra.OnInitialize(func() {
		// 打印 debug 日志
		// if globalReq.IsVerbose {
		// bdpan.SetLogLevel(logrus.DebugLevel)
		// }
	})
}
