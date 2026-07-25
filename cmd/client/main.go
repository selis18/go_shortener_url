package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

type UrlReq struct {
	Url string
}

func main() {
	endpoint := "http://localhost:8080/"

	data := url.Values{}
	fmt.Println("Введите длинный URL")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSpace(long)
	data.Set("longUrl", long)
	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(strings.NewReader(data.Encode())).
		Post(endpoint)

	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", resp.Status())
	fmt.Println(resp.String())
}
