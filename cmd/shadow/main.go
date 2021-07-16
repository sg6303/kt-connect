package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/proxy/dnsserver"
	"github.com/alibaba/kt-connect/pkg/proxy/shadowsocks"
	"github.com/alibaba/kt-connect/pkg/proxy/socks"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	go socks.Start()
	go shadowsocks.Start()
	log.Info().Msg("shadow staring...")
	srv := dnsserver.NewDNSServerDefault()
	err := srv.ListenAndServe()
	if err != nil {
		log.Error().Msg(err.Error())
		panic(err.Error())
	}
}
