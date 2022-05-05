package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/Mrtoy/dtcg-server/service"
	"github.com/manifoldco/promptui"
)

const deckInfo = `["Exported from http://digimon.card.moe","ST3-07","ST3-07","ST3-07","ST3-12","ST3-12","ST3-12","ST3-12","BT1-063","BT1-063","BT1-052","BT1-052","BT1-052","BT1-052","BT1-048","BT1-048","BT1-048","BT1-048","BT1-005","BT1-006","BT1-006","BT1-006","BT1-006","BT1-057","BT1-057","BT1-102","BT1-102","BT2-033","BT2-033","BT2-033","BT2-099","BT2-038","BT2-038","BT2-038","BT2-038","BT2-034","BT2-034","BT2-034","BT2-039","BT2-039","BT2-098","BT2-098","BT1-060","BT1-060","BT1-060","BT1-060","BT1-087","BT1-087","BT1-087","BT2-041","BT2-041","BT2-041","BT2-041","BT2-087","BT2-087","BT2-087"]`

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:2333")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	msgChan := make(chan *service.Event)
	go func() {
		reader := bufio.NewReader(conn)
		for {
			data, err := reader.ReadSlice('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			fmt.Println(string(data))
		}
	}()
	go func() {
		for {
			e := <-msgChan
			buf, err := json.Marshal(e)
			if err != nil {
				log.Println("Marshal failed,err:", err)
				continue
			}
			conn.Write(buf)
			conn.Write([]byte("\n"))
		}
	}()
	runCmd(msgChan, "room:join")
	runCmd(msgChan, "room:ready")
	for {
		prompt := promptui.Prompt{
			Label: ">",
		}
		cmd, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		exit := runCmd(msgChan, cmd)
		if exit {
			return
		}
	}
}

func runCmd(msgChan chan *service.Event, cmd string) (exit bool) {
	if cmd == "exit" {
		exit = true
		return
	} else if cmd == "room:join" {
		msgChan <- &service.Event{
			Type: cmd,
			Data: "test",
		}
	} else if cmd == "room:ready" {
		var deck []service.Card
		json.Unmarshal([]byte(deckInfo), &deck)
		deck = deck[1:]
		msgChan <- &service.Event{
			Type: cmd,
			Data: struct {
				Ready bool           `json:"ready"`
				Deck  []service.Card `json:"deck"`
			}{
				Ready: true,
				Deck:  deck,
			},
		}
	} else {
		msgChan <- &service.Event{
			Type: cmd,
		}
	}
	return
}
