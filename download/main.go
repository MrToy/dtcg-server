package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Mrtoy/dtcg-server/service"
)

func download(u string, id string) {
	_, err := os.Stat("./img/" + id + ".jpg")
	if !os.IsNotExist(err) {
		return
	}
	link := fmt.Sprintf("https://dtcg-assets.moecard.cn/img/%s~card", u)
	resp, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	f, err := os.Create("./img/" + id + ".jpg")
	if err != nil {
		log.Println(err)
		return
	}
	io.Copy(f, resp.Body)
	fmt.Println(id)
}

func main() {
	var wg sync.WaitGroup
	for _, card := range service.GetCardDetails() {
		p := card.Images[0].ImgPath
		if p == "" {
			p = card.Images[0].ThumbPath
		}
		wg.Add(1)
		go func(card *service.CardDetail) {
			download(p, card.Serial)
			wg.Done()
		}(card)
	}
	wg.Wait()
	// download("card/1873_2975.MnXUeFWUhtz.jpg", "1873_2975")
}
