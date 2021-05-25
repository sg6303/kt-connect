package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/entity"
	"github.com/lextoumbourou/goodhosts"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var interrupt = make(chan bool)

// StopBackendProcess ...  停止背后进程
func StopBackendProcess(stop bool, cancel func()) {
	if cancel == nil {
		return
	}
	cancel()
	interrupt <- stop
}

// Interrupt ...
func Interrupt() chan bool {
	return interrupt
}

// IsDaemonRunning check daemon is running or not
func IsDaemonRunning(pidFile string) bool {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// HomeDir Current User home dir
func HomeDir() string {
	// linux & mac
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	// windows
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	return "/root"
}

// KubeConfig location of kube-config file
func KubeConfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		kubeconfig = filepath.Join(HomeDir(), ".kube", "config")
	}
	return kubeconfig
}

// CreateDirIfNotExist create dir
func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

// WritePidFile write pid to file
func WritePidFile(pidFile string) (pid int, err error) {
	pid = os.Getpid()
	err = ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	return
}

// IsWindows check runtime is windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// DropHosts ... 删除host的映射关系
func DropHosts(hostsMap map[string]string) {
	hosts, err := goodhosts.NewHosts()

	if err != nil {
		log.Warn().Msgf("Fail to read hosts from host %s, ignore", err.Error())
		return
	}

	for name, ip := range hostsMap {
		if hosts.Has(ip, name) {
			if err = hosts.Remove(ip, name); err != nil {
				log.Warn().Str("ip", ip).Str("name", name).Msg("remove host failed")
			}
		}
	}

	if err := hosts.Flush(); err != nil {
		log.Error().Err(err).Msgf("Error Happen when flush hosts")
	}

	log.Info().Msgf("- drop hosts successful.")
}

// DumpHosts DumpToHosts  添加到hosts里面
func DumpHosts(hostsMap map[string]string) {
	hosts, err := goodhosts.NewHosts()

	if err != nil {
		log.Warn().Msgf("Fail to read hosts from host %s, ignore", err.Error())
		return
	}

	for name, ip := range hostsMap {
		if !hosts.Has(ip, name) {
			if err = hosts.Add(ip, name); err != nil {
				log.Warn().Str("ip", ip).Str("name", name).Msg("add host failed")
			}
		}
	}

	if err := hosts.Flush(); err != nil {
		log.Error().Err(err).Msg("Error Happen when dump hosts")
	}

	log.Info().Msg("Dump hosts successful.")

}

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func RegisterToConsul(hostsMap map[string]string, consulAddress string) {
	for name, ip := range hostsMap {
		log.Info().Msgf("Service: %s , ip: %s ===> now register to consul", name, ip)

		//如果consul上已经存在，就不注册，如果不存在，则直接注册
		res, err := http.Get("http://" + consulAddress + "/v1/catalog/service/" + name)
		perror(err)
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err.Error())
		}
		var data []entity.Server
		errJson := json.Unmarshal(body, &data)
		if errJson != nil {
			panic(errJson.Error())
		}
		fmt.Printf("Results: %v\n", data)

		if data != nil && len(data) != 0 {
			var server = data[0]
			if server.ServiceAddress == ip {
				//如果一致，就是已经存在
				continue
			} else {
				//将原来的删除
				req, err := http.NewRequest(http.MethodPut, "http://"+consulAddress+"/v1/agent/service/deregister/"+server.ServiceID, nil)
				if err != nil {
					continue
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Error().Msg("http post request consul is error")
				}
				defer resp.Body.Close()
			}
		}

		//发起注册
		addServer := entity.RegisterServer{ID: name + "-1", Name: name, Address: ip, Port: 80}
		postBody, err := json.Marshal(addServer)
		if err != nil {
			continue
		}
		responseBody := bytes.NewBuffer(postBody)
		req, err := http.NewRequest(http.MethodPut, "http://"+consulAddress+"/v1/agent/service/register", responseBody)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error().Msg("http post request consul is error")
		}
		defer resp.Body.Close()
	}

	log.Info().Msg(" RegisterToConsul successful.")
}

//从consul 下线
func DeregisterFromConsul(hostsMap map[string]string, consulAddress string) {
	for name, ip := range hostsMap {
		log.Info().Msgf("Service: %s , ip: %s ===> now deregister from consul", name, ip)

		//如果consul上已经存在，就注销
		res, err := http.Get("http://" + consulAddress + "/v1/catalog/service/" + name)
		perror(err)
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err.Error())
		}
		var data []entity.Server
		json.Unmarshal(body, &data)
		fmt.Printf("Results: %v\n", data)

		if data != nil && len(data) != 0 {
			var server = data[0]
			//将原来的删除
			req, err := http.NewRequest(http.MethodPut, "http://"+consulAddress+"/v1/agent/service/deregister/"+server.ServiceID, nil)
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Error().Msg("http post request consul is error")
			}
			defer resp.Body.Close()
		}
	}

	log.Info().Msg(" DeregisterFromConsul successful.")
}
