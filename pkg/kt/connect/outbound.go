package connect

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/channel"

	"github.com/alibaba/kt-connect/pkg/kt/options"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Outbound start vpn connection  开始vpn网络连接
func (s *Shadow) Outbound(name, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface) (err error) {
	ssh := channel.SSHChannel{}
	return outbound(s, name, podIP, credential, cidrs, cli, &ssh)
}

func outbound(s *Shadow, name, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface, ssh channel.Channel) (err error) {
	options := s.Options
	stop := make(chan bool)
	rootCtx, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		util.StopBackendProcess(<-stop, cancel)
	}()

	var wg sync.WaitGroup //定义一个等待组
	wg.Add(1)             //添加几个子任务
	go func(wg *sync.WaitGroup) {
		err = exec.BackgroundRunWithCtx(
			&exec.CMDContext{
				Ctx: rootCtx,
				Cmd: cli.Kubectl().PortForward(
					options.Namespace,
					name,
					options.ConnectOptions.SSHPort), //组装 kubectl port-forward 命令
				Name: "port-forward",
				Stop: stop,
			},
			options.Debug,
		)
		log.Info().Msgf("wait(%ds) port-forward successful", options.WaitTime)
		time.Sleep(time.Duration(options.WaitTime) * time.Second)
		wg.Done()
	}(&wg)

	wg.Wait() //只有上面的任务成功了，才可以 继续往下走
	if err != nil {
		return
	}

	if options.ConnectOptions.Method == "socks5" {
		err = startSocks5Connection(ssh, options)
	} else {
		err = startVPNConnection(rootCtx, cli, credential, options, podIP, cidrs, stop)
	}
	if err != nil {
		return
	}

	log.Info().Msgf("KT proxy start successful")
	return
}

//socks5 方式连接
func startSocks5Connection(ssh channel.Channel, options *options.DaemonOptions) (err error) {
	log.Info().Msgf("==============================================================")
	log.Info().Msgf("Start SOCKS5 Proxy Successful: export http_proxy=socks5://127.0.0.1:%d", options.ConnectOptions.Socke5Proxy)
	log.Info().Msgf("==============================================================")
	_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
		options.ConnectOptions.Socke5Proxy)), 0644)
	_ = ioutil.WriteFile(".envrc", []byte(fmt.Sprintf("KUBERNETES_NAMESPACE=%s",
		options.Namespace)), 0644)

	return ssh.StartSocks5Proxy(
		&channel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.Socke5Proxy),
	)
}

func startVPNConnection(rootCtx context.Context, cli exec.CliInterface, credential *util.SSHCredential,
	options *options.DaemonOptions, podIP string, cidrs []string, stop chan bool) (err error) {
	err = exec.BackgroundRunWithCtx(
		&exec.CMDContext{
			Ctx: rootCtx,
			Cmd: cli.SSHUttle().Connect(credential.RemoteHost, credential.PrivateKeyPath, options.ConnectOptions.SSHPort,
				podIP, options.ConnectOptions.DisableDNS, cidrs, options.Debug),
			Name: "vpn(sshuttle)",
			Stop: stop,
		},
		options.Debug,
	)
	return err
}
