package ltdapi

import (
	"fmt"
	"net/http"
	"os"
)

const url = "https://apiv2.legiontd2.com/games?limit=50&sortBy=date&sortDirection=1&includeDetails=true&dateAfter=2022-07-01&offset=%v"

type LtdApi struct {
	Key string
}

type LTDResponse struct {
	Games []Game
}

type Game struct {
	PlayersData []PlayersData
}

type PlayersData struct {
	Cross                      bool
	MercenariesReceivedPerWave [][]string
	LeaksPerWave               [][]string
	BuildPerWave               [][]string
}

func New() *LtdApi {
	key := os.Getenv("apikey")
	return &LtdApi{
		Key: key,
	}
}

func (api *LtdApi) Request(offset int) (*http.Response, error) {
	pUrl := fmt.Sprintf(url, offset)
	req, err := http.NewRequest("GET", pUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", api.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	return resp, err
}
