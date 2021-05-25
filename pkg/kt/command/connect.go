package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
)

// newConnectCommand return new connect command
func newConnectCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "connect",
		Usage: "connection to kubernetes cluster",
		Flags: ConnectActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				//设置日志级别
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			return action.Connect(cli, options)
		},
	}
}

// Connect connect vpn to kubernetes cluster  连接k8s集群
func (action *Action) Connect(cli kt.CliInterface, options *options.DaemonOptions) (err error) {
	//利用pid文件判断是否已经在运行，如果是，则报错退出。
	if util.IsDaemonRunning(options.RuntimeOptions.PidFile) {
		return fmt.Errorf("connect already running %s exit this", options.RuntimeOptions.PidFile)
	}

	//声明关闭的处理
	ch := SetUpCloseHandler(cli, options, "connect")

	//连接集群
	if err = connectToCluster(cli, options); err != nil {
		return
	}

	//退出執行
	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-util.Interrupt()
		CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return
}

//连接集群
func connectToCluster(cli kt.CliInterface, options *options.DaemonOptions) (err error) {

	//获取pid，并写入到pid文件中，用于上面判断
	pid, err := util.WritePidFile(options.RuntimeOptions.PidFile)
	if err != nil {
		return
	}
	log.Info().Msgf("Connect Start At %d", pid)

	//获取要连接的集群信息
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return
	}

	//根據 命令參數判斷是否需要dump到hosts，修改hosts内容
	if options.ConnectOptions.Dump2Hosts {
		setupDump2Host(options, kubernetes)
	}

	if options.ConnectOptions.Consul {
		//需要注册到 consul
		registerToConsul(options, kubernetes)
	}

	//获取或创建影子实例
	endPointIP, podName, credential, err := getOrCreateShadow(options, err, kubernetes)
	if err != nil {
		return
	}

	//获取集群的cidr列表
	cidrs, err := kubernetes.ClusterCrids(options.Namespace, options.ConnectOptions.CIDR)
	if err != nil {
		return
	}

	return cli.Shadow().Outbound(podName, endPointIP, credential, cidrs, cli.Exec())
}

//创建影子实例
func getOrCreateShadow(options *options.DaemonOptions, err error, kubernetes cluster.KubernetesInterface) (string, string, *util.SSHCredential, error) {
	workload := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	if options.ConnectOptions.ShareShadow {
		workload = fmt.Sprintf("kt-connect-daemon-connect-shared")
	}

	endPointIP, podName, sshcm, credential, err :=
		kubernetes.GetOrCreateShadow(workload, options.Namespace, options.Image, labels(workload, options), envs(options),
			options.Debug, options.ConnectOptions.ShareShadow)
	if err != nil {
		return "", "", nil, err
	}

	// record shadow name will clean up terminal
	options.RuntimeOptions.Shadow = workload
	options.RuntimeOptions.SSHCM = sshcm

	return endPointIP, podName, credential, nil
}

//dump k8s 服务到 hosts
func setupDump2Host(options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	hosts := kubernetes.ServiceHosts(options.Namespace)
	for k, v := range hosts {
		log.Info().Msgf("Service found: %s %s", k, v)
	}
	if options.ConnectOptions.Dump2HostsNamespaces != nil {
		for _, namespace := range options.ConnectOptions.Dump2HostsNamespaces {
			if namespace == options.Namespace {
				continue
			}
			log.Debug().Msgf("Search service in %s namespace...", namespace)
			singleHosts := kubernetes.ServiceHosts(namespace)
			for k, v := range singleHosts {
				if v == "" || v == "None" {
					continue
				}
				log.Info().Msgf("Service found: %s.%s %s", k, namespace, v)
				hosts[k+"."+namespace] = v
			}
		}
	}
	util.DumpHosts(hosts)
	options.ConnectOptions.Hosts = hosts
}

//注册到consul
func registerToConsul(options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	hosts := kubernetes.ServiceHosts(options.Namespace)
	util.RegisterToConsul(hosts, options.ConnectOptions.ConsulAddress)
	options.ConnectOptions.ConsulServers = hosts
}

//环境变量
func envs(options *options.DaemonOptions) map[string]string {
	envs := make(map[string]string)
	if options.ConnectOptions.LocalDomain != "" {
		envs[common.EnvVarLocalDomain] = options.ConnectOptions.LocalDomain
	}
	return envs
}

//标签
func labels(workload string, options *options.DaemonOptions) map[string]string {
	labels := map[string]string{
		"kt-component": "connect",
		"control-by":   "kt",
	}
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}
	splits := strings.Split(workload, "-")
	labels["version"] = splits[len(splits)-1]
	return labels
}
