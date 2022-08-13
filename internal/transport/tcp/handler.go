package handler

import (
	"encoding/binary"
	"fmt"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/internal/domain"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/internal/service"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net"
	"time"
)

type Handler struct {
	rg *service.RequestGenerator
}

func NewHandler(RequestGenerator *service.RequestGenerator) *Handler {
	return &Handler{rg: RequestGenerator}
}

func getBsonBytesLength(lengthSlice []byte) int {
	return int(binary.LittleEndian.Uint32(lengthSlice))
}

func (h *Handler) readConnection(conn net.Conn, resMsgChan chan<- domain.ConBytesMsg, ttl time.Duration) error {
	for {
		ts := time.Now()
		fullBody := make([]byte, 4)
		_, err := conn.Read(fullBody)
		lengthBytes := getBsonBytesLength(fullBody)
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
	for {
		msg := <-reqMsgChan
		_, err := conn.Write(msg.MsgBody)
		if err != nil {
			return fmt.Errorf("error occured while write responce from router to tcp conn: %s", err.Error())
		}
	}
}

func (h *Handler) WriteResponseStat(resMsgChan chan domain.ConBytesMsg, stopMainChan chan bool) {
	for {
		resBytes := <-resMsgChan
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
		h.rg.PrintStatReport()
	}
}
