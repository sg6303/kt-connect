package channel

import (
	"fmt"
	"io"
	"net"

	"github.com/armon/go-socks5"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"

	"golang.org/x/crypto/ssh"
)

// SSHChannel ssh channel
type SSHChannel struct{}

// StartSocks5Proxy start socks5 proxy  开始 socks5 代理
func (c *SSHChannel) StartSocks5Proxy(certificate *Certificate, sshAddress string, socks5Address string) (err error) {
	//使用 root:root 用户登录连接
	conn, err := connection(certificate.Username, certificate.Password, sshAddress)
	defer conn.Close()
	if err != nil {
		return err
	}

	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return conn.Dial(network, addr)
		},
	}

	serverSocks, err := socks5.New(conf)

	if err != nil {
		return err
	}

	if err := serverSocks.ListenAndServe("tcp", socks5Address); err != nil {
		fmt.Println("failed to create socks5 server", err)
		return err
	}
	fmt.Println("dynamic port forward successful")
	return nil
}

// ForwardRemoteToLocal forward remote request to local
func (c *SSHChannel) ForwardRemoteToLocal(certificate *Certificate, sshAddress string, remoteEndpoint string, localEndpoint string) (err error) {
	conn, err := connection(certificate.Username, certificate.Password, sshAddress)
	defer conn.Close()
	if err != nil {
		log.Error().Msgf("fail to create ssh tunnel")
		return err
	}

	// Listen on remote server port
	listener, err := conn.Listen("tcp", remoteEndpoint)
	if err != nil {
		log.Error().Msgf("fail to listen remote endpoint ")
		return err
	}
	defer listener.Close()

	log.Info().Msgf("forward %s to localEndpoint %s", remoteEndpoint, localEndpoint)

	// handle incoming connections on reverse forwarded tunnel
	for {
		// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
		local, err := net.Dial("tcp", localEndpoint)
		if err != nil {
			log.Error().Msgf("Dial INTO local service error: %s", err)
			return err
		}

		client, err := listener.Accept()
		if err != nil {
			log.Error().Msgf("error: %s", err)
			return err
		}

		handleClient(client, local)
	}
}

func connection(username string, password string, address string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.Auth = []ssh.AuthMethod{
		ssh.Password(password),
	}

	//ssh连接到远程 address 地址
	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Error().Msgf("fail create ssh connection %s", err)
	}
	return conn, err
}

func handleClient(client net.Conn, remote net.Conn) {
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Error().Msgf("error while copy remote->local: %s", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Error().Msgf("error while copy local->remote: %s", err)
		}
		chDone <- true
	}()

	<-chDone
}
