package handler

import (
	"fmt"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/internal/domain"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net"
	"time"
)

type Handler struct {
	rg RGenerator
}

type RGenerator interface {
	GetSent() int
	GetReceived() int
	GetFinished() bool
	GetDurations() []time.Duration
	SetSent()
	SetReceived()
	SetFinished()
	SetDuration(duration time.Duration)
	GenerateRequestsEndless(reqBytesChan chan<- domain.ByteMsg)
	PrintStatReport()
}

func NewHandler(rg RGenerator) *Handler {
	return &Handler{rg: rg}
}

func (h *Handler) readConnection(conn net.Conn, resMsgChan chan<- domain.ConBytesMsg, ttl time.Duration) error {
	for {
		ts := time.Now()
		fullBody := make([]byte, 4)
		_, err := conn.Read(fullBody)
		lengthBytes := utils.GetBsonBytesLength(fullBody)
		if err != nil {
			return err
		}

		if lengthBytes > 0 {
			for {
				if len(fullBody) == lengthBytes {
					resMsgChan <- domain.ConBytesMsg{
						Conn:    conn,
						MsgBody: fullBody}
					break
				}
				if time.Since(ts) > ttl {
					return fmt.Errorf("Message read ttl %s exceeded time sinse %s\n", ttl, time.Since(ts))
				}

				restBytes := lengthBytes - len(fullBody)
				restBody := make([]byte, restBytes)
				receivedBytes, err := conn.Read(restBody)
				if err != nil {
					return err
				}

				if receivedBytes > 0 {
					fullBody = append(fullBody, restBody[:receivedBytes]...)
				}
			}
		}
	}
}

func (h *Handler) writeConnection(conn net.Conn, reqMsgChan <-chan domain.ByteMsg) error {
	for msg := range reqMsgChan {
		_, err := conn.Write(msg.MsgBody)
		if err != nil {
			return fmt.Errorf("error occured while write responce from router to tcp conn: %s", err.Error())
		}
	}
	return nil
}

func (h *Handler) WriteResponseStat(resMsgChan <-chan domain.ConBytesMsg, stopMainChan chan bool) {
	logrus.Info("WriteResponseStat")
	for resBytes := range resMsgChan {
		var res domain.RouteResponse
		err := bson.Unmarshal(resBytes.MsgBody, &res)
		if err != nil {
			logrus.Errorf("can't unmarshal bson response err = %s", err)
		}
		duration := time.Since(res.SendTime)
		h.rg.SetReceived()
		h.rg.SetDuration(duration)

		if h.rg.GetFinished() {
			if h.rg.GetSent() == h.rg.GetReceived() {
				stopMainChan <- true
			}
		}
	}
}

func (h *Handler) GenerateAndSend(conn net.Conn) {
	reqBytesChan := make(chan domain.ByteMsg, 100)
	resMsgChan := make(chan domain.ConBytesMsg, 100)
	stopMainChan := make(chan bool)

	go h.rg.GenerateRequestsEndless(reqBytesChan)

	for i := 0; i < 100; i++ {
		go h.WriteResponseStat(resMsgChan, stopMainChan)
	}

	go func() {
		err := h.writeConnection(conn, reqBytesChan)
		if err != nil {
			logrus.Errorf("writeConnection error %s", err)
			stopMainChan <- true
			return
		}
	}()
	go func() {
		err := h.readConnection(conn, resMsgChan, 2*time.Second)
		if err != nil {
			logrus.Errorf("readConnection error %s", err)
			stopMainChan <- true
			return
		}
	}()

	if <-stopMainChan {
		if err := conn.Close(); err != nil {
			logrus.WithFields(
				logrus.Fields{
					"process": "close connection on stop main chan",
				}).Error(err)
		}
		close(reqBytesChan)
		close(resMsgChan)
		close(stopMainChan)
		h.rg.PrintStatReport()
	}
}
