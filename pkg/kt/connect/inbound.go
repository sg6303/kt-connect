package connect

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/channel"
	"github.com/alibaba/kt-connect/pkg/kt/options"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Inbound mapping local port from cluster  远程->本地
func (s *Shadow) Inbound(exposePorts, podName, remoteIP string, credential *util.SSHCredential) (err error) {
	kubernetesCli := &kubectl.Cli{KubeOptions: s.Options.KubeOptions}
	ssh := &channel.SSHChannel{}
	log.Info().Msg("creating shadow inbound(remote->local)")
	return inbound(exposePorts, podName, remoteIP, credential, s.Options, kubernetesCli, ssh)
}

func inbound(exposePorts, podName, remoteIP string, credential *util.SSHCredential,
	options *options.DaemonOptions,
	kubernetesCli kubectl.CliInterface,
	ssh channel.Channel,
) (err error) {
	stop := make(chan bool)
	rootCtx, cancel := context.WithCancel(context.Background())

	// one of the background process start failed and will cancel the started process
	go func() {
		util.StopBackendProcess(<-stop, cancel)
	}()

	//生成本地ssh端口
	log.Info().Msgf("remote %s forward to local %s", remoteIP, exposePorts)
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}
	err = portForward(rootCtx, kubernetesCli, podName, localSSHPort, stop, options)

	if err != nil {
		return
	}

	exposeLocalPortsToRemote(ssh, exposePorts, localSSHPort)
	return nil
}

func portForward(rootCtx context.Context, kubernetesCli kubectl.CliInterface, podName string, localSSHPort int,
	stop chan bool, options *options.DaemonOptions,
) error {
	var err error
	var wg sync.WaitGroup

	debug := options.Debug
	namespace := options.Namespace

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		portforward := kubernetesCli.PortForward(namespace, podName, localSSHPort)
		err = exec.BackgroundRunWithCtx(
			&exec.CMDContext{
				Ctx:  rootCtx,
				Cmd:  portforward,
				Name: "exchange port forward to local",
				Stop: stop,
			},
			debug,
		)
		log.Info().Msgf("wait(%ds) port-forward successful", options.WaitTime)
		time.Sleep(time.Duration(options.WaitTime) * time.Second)
		wg.Done()
	}(&wg)
	wg.Wait()
	return err
}

func exposeLocalPortsToRemote(ssh channel.Channel, exposePorts string, localSSHPort int) {
	var wg sync.WaitGroup
	// supports multi port pairs
	portPairs := strings.Split(exposePorts, ",")
	for _, exposePort := range portPairs {
		localPort, remotePort := getPortMapping(exposePort)
		exposeLocalPortToRemote(wg, ssh, remotePort, localPort, localSSHPort)
	}
	wg.Wait()
}

func exposeLocalPortToRemote(wg sync.WaitGroup, ssh channel.Channel, remotePort string, localPort string, localSSHPort int) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info().Msgf("exposeLocalPortsToRemote request from pod:%s to 127.0.0.1:%s\n", remotePort, localPort)
		err := ssh.ForwardRemoteToLocal(
			&channel.Certificate{
				Username: "root",
				Password: "root",
			},
			fmt.Sprintf("127.0.0.1:%d", localSSHPort),
			fmt.Sprintf("0.0.0.0:%s", remotePort),
			fmt.Sprintf("127.0.0.1:%s", localPort),
		)
		if err != nil {
			log.Error().Msgf("error happen when forward remote request to local %s", err)
		}
		log.Info().Msgf("exposeLocalPortsToRemote request from pod:%s to 127.0.0.1:%s finished\n", remotePort, localPort)
		wg.Done()
	}(&wg)
}

func getPortMapping(exposePort string) (string, string) {
	localPort := exposePort
	remotePort := exposePort
	ports := strings.SplitN(exposePort, ":", 2)
	if len(ports) > 1 {
		localPort = ports[1]
		remotePort = ports[0]
	}
	return localPort, remotePort
}
