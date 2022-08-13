package app

import (
	"github.com/VadimGossip/tcpBsonServerReqGenerator/internal/config"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/internal/service"
	handler "github.com/VadimGossip/tcpBsonServerReqGenerator/internal/transport/tcp"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

func Run(configDir string) {
	cfg, err := config.Init(configDir)
	if err != nil {
		logrus.Errorf("Config initialization error %s", err)
	}
	conn, err := net.Dial("tcp", cfg.ServerListenerTcp.Host+":"+strconv.Itoa(cfg.ServerListenerTcp.Port))
	defer conn.Close()
	if err != nil {
		logrus.Fatalf("error occured while connecting to tcp server: %s", err.Error())
	}
	rg := service.NewRequestGenerator(cfg.RGenerator)
	hlr := handler.NewHandler(rg)
	hlr.GenerateAndSend(conn)
}
