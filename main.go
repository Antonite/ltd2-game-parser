package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/antonite/ltd2-game-parser/ltdapi"
)

const gameOffset = 50000

var trash = []string{"golem_unit_id", "mudman_unit_id", "infiltrator_unit_id", "orchid_unit_id", "kingpin_unit_id", "sakura_unit_id", "veteran_unit_id", "peewee_unit_id", "nekomata_unit_id"}

func main() {
	api := ltdapi.New()
	// if err := generateUnits(api); err != nil {
	// 	panic(err)
	// }

	// if err := generateWaves(api); err != nil {
	// 	panic(err)
	// }

	if err := generateData(api); err != nil {
		panic(err)
	}

}

func generateData(api *ltdapi.LtdApi) error {
	csvFile, err := os.Create("data.csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()

	widths := []string{"0.5", "1", "1.5", "2", "2.5", "3", "3.5", "4", "4.5", "5", "5.5", "6", "6.5", "7", "7.5", "8", "8.5"}
	heights := []string{"0.5", "1", "1.5", "2", "2.5", "3", "3.5", "4", "4.5", "5", "5.5", "6", "6.5", "7", "7.5", "8", "8.5", "9", "9.5", "10", "10.5", "11", "11.5", "12", "12.5", "13", "13.5"}

	// define the board
	keys := []string{}
	for _, bw := range widths {
		for _, bh := range heights {
			keys = append(keys, bw+"|"+bh)
		}
	}
	sort.Strings(keys)

	// add headers
	data := []string{"leak", "wave", "sends"}
	data = append(data, keys...)
	w := csv.NewWriter(csvFile)
	w.Write(data)
	w.Flush()

	sendDic := make(map[string]bool)
	off := 0
	date := "2022-06-15T00:00:00.000Z"
	for off <= gameOffset {
		fmt.Printf("offset: %v\n", off)
		resp, err := api.Request(off, date)
		if err != nil || resp.StatusCode != 200 {
			return err
		}

		games := []ltdapi.Game{}
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&games); err != nil {
			panic(err)
		}

		var lastDate string
		sendDic, lastDate = processResp(games, keys, csvFile, sendDic)

		off += 50
		if gameOffset > 50000 {
			date = lastDate
			off = 0
		}
	}

	processSends(sendDic)

	return nil
}

func processResp(games []ltdapi.Game, keys []string, csvFile *os.File, sendDic map[string]bool) (map[string]bool, string) {
	w := csv.NewWriter(csvFile)
	defer w.Flush()

	for _, g := range games {
		for _, p := range g.PlayersData {
			if p.Cross == true || len(p.LeaksPerWave) < 10 {
				continue
			}

			allData := [][]string{}
			badUnitFound := false
			for i := 0; i < len(p.LeaksPerWave) && i < 10; i++ {
				data := []string{}

				// did we leak
				if len(p.LeaksPerWave[i]) == 0 {
					data = append(data, "0")
				} else {
					data = append(data, "1")
				}

				// append wave #
				data = append(data, strconv.Itoa(i+1))

				// sends received
				sort.Strings(p.MercenariesReceivedPerWave[i])
				sends := strings.Join(p.MercenariesReceivedPerWave[i], ",")
				sendDic[sends] = true
				data = append(data, sends)

				// units built
				m := make(map[string]string)
				for _, u := range p.BuildPerWave[i] {
					s := strings.Split(u, ":")
					// skip trash
					if isTrashUnit(s[0]) {
						badUnitFound = true
					}
					m[s[1]] = s[0]
				}

				if badUnitFound {
					break
				}

				for _, k := range keys {
					unit, ok := m[k]
					if !ok {
						data = append(data, "")
					} else {
						data = append(data, unit)
					}
				}

				allData = append(allData, data)
			}

			if !badUnitFound {
				w.WriteAll(allData)
			}
		}
	}

	return sendDic, games[len(games)-1].Date
}

func processSends(sendDic map[string]bool) {
	csvFile, err := os.Create("sends.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	w.Write([]string{"leak", "sends"})
	w.Write([]string{"0", ""})
	defer w.Flush()

	for key := range sendDic {
		w.Write([]string{"0", key})
	}
}

func generateUnits(api *ltdapi.LtdApi) error {
	csvFile, err := os.Create("units.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	w.Write([]string{"leak", "units"})
	w.Write([]string{"0", ""})
	w.Flush()

	off := 0
	for {
		fmt.Printf("offset: %v\n", off)
		resp, err := api.RequestUnits(off)
		if err != nil || resp.StatusCode != 200 {
			return err
		}

		units := []ltdapi.Unit{}
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&units); err != nil {
			return err
		}

		processUnits(units, csvFile)

		off += 50
	}
}

func processUnits(units []ltdapi.Unit, csvFile *os.File) {
	w := csv.NewWriter(csvFile)
	defer w.Flush()

	for _, u := range units {
		if u.UnitClass != "Fighter" || isTrashUnit(u.UnitId) {
			continue
		}

		w.Write([]string{"0", u.UnitId})
	}
}

func isTrashUnit(unit string) bool {
	for _, t := range trash {
		if t == unit {
			return true
		}
	}

	return false
}

func generateWaves(api *ltdapi.LtdApi) error {
	csvFile, err := os.Create("waves.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	defer w.Flush()

	w.Write([]string{"leak", "wave"})
	for i := 0; i < 11; i++ {
		w.Write([]string{"0", strconv.Itoa(i)})
	}

	return nil
}
