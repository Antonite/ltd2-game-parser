package ltdapi

import (
	"fmt"
	"net/http"
	"os"
)

const url = "https://apiv2.legiontd2.com/games?limit=50&sortBy=date&sortDirection=1&includeDetails=true&dateAfter=%v&offset=%v"
const unitsUrl = "https://apiv2.legiontd2.com/units/byVersion/9.05.6?limit=50&enabled=true&offset=%v"

type LtdApi struct {
	Key string
}

type LTDResponse struct {
	Games []Game
}

type Game struct {
	PlayersData []PlayersData
	Date        string
}

type PlayersData struct {
	Cross                      bool
	MercenariesReceivedPerWave [][]string
	LeaksPerWave               [][]string
	BuildPerWave               [][]string
}

type Unit struct {
	UnitId    string
	UnitClass string
}

func New() *LtdApi {
	key := os.Getenv("apikey")
	return &LtdApi{
		Key: key,
	}
}

func (api *LtdApi) Request(offset int, date string) (*http.Response, error) {
	pUrl := fmt.Sprintf(url, date, offset)
	req, err := http.NewRequest("GET", pUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", api.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	return resp, err
}

func (api *LtdApi) RequestUnits(offset int) (*http.Response, error) {
	pUrl := fmt.Sprintf(unitsUrl, offset)
	req, err := http.NewRequest("GET", pUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", api.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	return resp, err
}
