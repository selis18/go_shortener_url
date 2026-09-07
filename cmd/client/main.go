package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

type URLReq struct {
	URL string
}

func main() {
	endpoint := "http://localhost:8080/api/shorten"

	fmt.Println("Введите длинный URL")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSpace(long)

	client := resty.New()
	client.SetDebug(true)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"url": long}).
		Post(endpoint)

	fmt.Printf("Отправляем: %q\n", long)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", resp.Status())
	fmt.Println("Заголовки ответа:", resp.Header())
	fmt.Println(resp.String())
}
