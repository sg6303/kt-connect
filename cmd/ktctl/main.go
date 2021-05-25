package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/kt/util"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // 当导入一个包时，该包下的文件里所有init()函数都会被执行，然而，有些时候我们并不需要把整个包都导入进来，仅仅是是希望它执行init()函数而已
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	version = "0.0.13-rc9"
)

func init() {
	//设置全局日志级别 info
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	//输出格式
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: util.IsWindows()})
}

//先执行 import 的 init ，然后再执行 当前的 init 方法。最后到 main
func main() {
	// 得到默认 cli 操作实体
	options := opt.NewDaemonOptions()

	// 创建 cli 应用
	app := cli.NewApp()
	app.Name = "KT Connect"
	app.Usage = ""
	app.Version = version
	app.Authors = command.NewCliAuthor()
	app.Flags = command.AppFlags(options, version)

	context := &kt.Cli{Options: options}
	action := command.Action{}

	app.Commands = command.NewCommands(context, &action, options)
	err := app.Run(os.Args)
	if err != nil {
		log.Error().Msg(err.Error())
		command.CleanupWorkspace(context, options)
		os.Exit(-1)
	}
}
