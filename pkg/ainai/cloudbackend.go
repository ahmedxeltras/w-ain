// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package ainai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wavetermdev/ainterm/pkg/aincloud"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/panichandler"
)

type WaveAICloudBackend struct{}

var _ AIBackend = WaveAICloudBackend{}

const CloudWebsocketConnectTimeout = 1 * time.Minute
const OpenAICloudReqStr = "openai-cloudreq"
const PacketEOFStr = "EOF"

type WaveAICloudReqPacketType struct {
	Type       string                             `json:"type"`
	ClientId   string                             `json:"clientid"`
	Prompt     []ainshrpc.WaveAIPromptMessageType `json:"prompt"`
	MaxTokens  int                                `json:"maxtokens,omitempty"`
	MaxChoices int                                `json:"maxchoices,omitempty"`
}

func MakeWaveAICloudReqPacket() *WaveAICloudReqPacketType {
	return &WaveAICloudReqPacketType{
		Type: OpenAICloudReqStr,
	}
}

func (WaveAICloudBackend) StreamCompletion(ctx context.Context, request ainshrpc.WaveAIStreamRequest) chan ainshrpc.RespOrErrorUnion[ainshrpc.WaveAIPacketType] {
	rtn := make(chan ainshrpc.RespOrErrorUnion[ainshrpc.WaveAIPacketType])
	wsEndpoint := aincloud.GetWSEndpoint()
	go func() {
		defer func() {
			panicErr := panichandler.PanicHandler("WaveAICloudBackend.StreamCompletion", recover())
			if panicErr != nil {
				rtn <- makeAIError(panicErr)
			}
			close(rtn)
		}()
		if wsEndpoint == "" {
			rtn <- makeAIError(fmt.Errorf("no cloud ws endpoint found"))
			return
		}
		if request.Opts == nil {
			rtn <- makeAIError(fmt.Errorf("no openai opts found"))
			return
		}
		websocketContext, dialCancelFn := context.WithTimeout(context.Background(), CloudWebsocketConnectTimeout)
		defer dialCancelFn()
		conn, _, err := websocket.DefaultDialer.DialContext(websocketContext, wsEndpoint, nil)
		if err == context.DeadlineExceeded {
			rtn <- makeAIError(fmt.Errorf("OpenAI request, timed out connecting to cloud server: %v", err))
			return
		} else if err != nil {
			rtn <- makeAIError(fmt.Errorf("OpenAI request, websocket connect error: %v", err))
			return
		}
		defer func() {
			err = conn.Close()
			if err != nil {
				rtn <- makeAIError(fmt.Errorf("unable to close openai channel: %v", err))
			}
		}()
		var sendablePromptMsgs []ainshrpc.WaveAIPromptMessageType
		for _, promptMsg := range request.Prompt {
			if promptMsg.Role == "error" {
				continue
			}
			sendablePromptMsgs = append(sendablePromptMsgs, promptMsg)
		}
		reqPk := MakeWaveAICloudReqPacket()
		reqPk.ClientId = request.ClientId
		reqPk.Prompt = sendablePromptMsgs
		reqPk.MaxTokens = request.Opts.MaxTokens
		reqPk.MaxChoices = request.Opts.MaxChoices
		configMessageBuf, err := json.Marshal(reqPk)
		if err != nil {
			rtn <- makeAIError(fmt.Errorf("OpenAI request, packet marshal error: %v", err))
			return
		}
		err = conn.WriteMessage(websocket.TextMessage, configMessageBuf)
		if err != nil {
			rtn <- makeAIError(fmt.Errorf("OpenAI request, websocket write config error: %v", err))
			return
		}
		for {
			_, socketMessage, err := conn.ReadMessage()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("err received: %v", err)
				rtn <- makeAIError(fmt.Errorf("OpenAI request, websocket error reading message: %v", err))
				break
			}
			var streamResp *ainshrpc.WaveAIPacketType
			err = json.Unmarshal(socketMessage, &streamResp)
			if err != nil {
				rtn <- makeAIError(fmt.Errorf("OpenAI request, websocket response json decode error: %v", err))
				break
			}
			if streamResp.Error == PacketEOFStr {
				// got eof packet from socket
				break
			} else if streamResp.Error != "" {
				// use error from server directly
				rtn <- makeAIError(fmt.Errorf("%v", streamResp.Error))
				break
			}
			rtn <- ainshrpc.RespOrErrorUnion[ainshrpc.WaveAIPacketType]{Response: *streamResp}
		}
	}()
	return rtn
}
