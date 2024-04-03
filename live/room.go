package live

import (
	"bytes"
	"compress/gzip"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/yunbaifan/pkg/live/protobuf"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Room struct {
	// 房间 地址
	URL                  string
	TtwID                string
	RoomID               string
	RoomStore            string // 房间存储
	RoomTitle            string // 房间标题
	wsConn               *websocket.Conn
	Errch                chan ErrChannel
	WebcastChatMessage   chan []byte
	WebcastGiftMessage   chan []byte
	WebcastLikeMessage   chan []byte
	WebcastMemberMessage chan []byte
}

type ErrChannel struct {
	Err  error
	Type string
}

func (r *Room) Close() {
	_ = r.wsConn.Close()
}

func (r *Room) recover() {
	if err := recover(); err != nil {
		buf := make([]byte, 64<<10)
		n := runtime.Stack(buf, false)
		buf = buf[:n]
		log.Printf("write stop %v \n%s\n", err, buf)
		return
	}
}

func Start(r *Room, wsURI string) error {
	r.Errch = make(chan ErrChannel)
	r.WebcastChatMessage = make(chan []byte)
	r.WebcastGiftMessage = make(chan []byte)
	r.WebcastLikeMessage = make(chan []byte)
	r.WebcastMemberMessage = make(chan []byte)
	wsURI = strings.Replace(wsURI, "%s", r.RoomID, -1)
	h := http.Header{}
	h.Set("cookie", "ttwid="+r.TtwID)
	h.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	wsConn, wsResp, err := websocket.DefaultDialer.Dial(wsURI, h)
	if err != nil {
		return err
	}
	defer func() {
		_ = wsResp.Body.Close()
	}()
	log.Printf("连接房间:%s 连接房间状态: %d \n", r.RoomID, wsResp.StatusCode)
	r.wsConn = wsConn
	go r.ProcessData()
	go r.read()
	go r.send()

	return nil

}

func (r *Room) ProcessData() {
	for {
		select {
		case err := <-r.Errch:
			log.Printf("Error: %v, Type: %s", err.Err, err.Type)
		case msg, _ := <-r.WebcastChatMessage:
			var chatMsg protobuf.ChatMessage
			_ = proto.Unmarshal(msg, &chatMsg)
			log.Printf("[弹幕] %s : %s\n", chatMsg.User.NickName, chatMsg.Content)
		case msg, _ := <-r.WebcastGiftMessage:
			var giftMsg protobuf.GiftMessage
			_ = proto.Unmarshal(msg, &giftMsg)
			log.Printf("[礼物] %s %s \n", giftMsg.User.NickName, giftMsg.Gift.Name)
		case msg, _ := <-r.WebcastLikeMessage:
			var likeMsg protobuf.LikeMessage
			_ = proto.Unmarshal(msg, &likeMsg)
			log.Printf("[点赞] %s 点赞 * %d \n", likeMsg.User.NickName, likeMsg.Count)
		case memberMsg, _ := <-r.WebcastMemberMessage:
			var enterMsg protobuf.MemberMessage
			_ = proto.Unmarshal(memberMsg, &enterMsg)
			log.Printf("[入场] %s 直播间\n", enterMsg.User.NickName)
		}
	}
}

func (r *Room) send() {
	defer func() {
		recover()
	}()
	// 发送消息
	for {
		var (
			err  error
			data []byte
		)
		if data, err = proto.Marshal(&protobuf.PushFrame{
			PayloadType: "bh",
		}); err != nil {
			r.Errch <- ErrChannel{
				Err:  err,
				Type: "proto.Marshal",
			}
			continue
		}
		if err = r.wsConn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			r.Errch <- ErrChannel{
				Err:  err,
				Type: "writeMessage",
			}
			continue
		}
		// 心跳包
		time.Sleep(time.Second * 10)
	}
}

func (r *Room) read() {
	defer r.recover()
	for {
		_, message, err := r.wsConn.ReadMessage()
		if err != nil {
			r.Errch <- ErrChannel{
				Err:  err,
				Type: "readMessage",
			}
			continue
		}
		var msgPack protobuf.PushFrame
		_ = proto.Unmarshal(message, &msgPack)
		decompressed, _ := r.deGzip(msgPack.Payload)
		var payloadPackage protobuf.Response
		_ = proto.Unmarshal(decompressed, &payloadPackage)
		if payloadPackage.NeedAck {
			r.sendAck(msgPack.LogId, payloadPackage.InternalExt)
		}
		for _, msg := range payloadPackage.MessagesList {
			switch msg.Method {
			case "WebcastChatMessage":
				r.WebcastChatMessage <- msg.Payload
			case "WebcastGiftMessage":
				r.WebcastGiftMessage <- msg.Payload
			case "WebcastLikeMessage":
				r.WebcastLikeMessage <- msg.Payload
			case "WebcastMemberMessage":
				r.WebcastMemberMessage <- msg.Payload
			}
		}
	}
}
func (r *Room) deGzip(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	var out bytes.Buffer
	gr, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&out, gr)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (r *Room) sendAck(logID uint64, iExt string) {
	ackPack := &protobuf.PushFrame{
		LogId:       logID,
		PayloadType: iExt,
	}
	data, _ := proto.Marshal(ackPack)
	if err := r.wsConn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		r.Errch <- ErrChannel{
			Err:  err,
			Type: "writeMessage",
		}
	}
}
