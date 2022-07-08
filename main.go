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

var trash = []string{"golem_unit_id", "mudman_unit_id", "infiltrator_unit_id", "orchid_unit_id", "kingpin_unit_id", "sakura_unit_id", "veteran_unit_id", "peewee_unit_id", "nekomata_unit_id"}

func main() {
	api := ltdapi.New()

	csvFile, err := os.Create("data.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
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

	off := 0
	for off <= 5000 {
		fmt.Printf("offset: %v\n", off)
		resp, err := api.Request(off)
		if err != nil {
			panic(err)
		}

		games := []ltdapi.Game{}
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&games); err != nil {
			panic(err)
		}

		processResp(games, keys, csvFile)

		off += 50
	}
}

func processResp(games []ltdapi.Game, keys []string, csvFile *os.File) {
	w := csv.NewWriter(csvFile)
	defer w.Flush()
	for _, g := range games {
		for _, p := range g.PlayersData {
			if p.Cross == true {
				continue
			}

			badUnitFound := false
			for i := 0; i < len(p.LeaksPerWave) && i < 10 && !badUnitFound; i++ {
				data := []string{}

				// append wave #
				data = append(data, strconv.Itoa(i+1))

				// did we leak
				if len(p.LeaksPerWave[i]) == 0 {
					data = append(data, "0")
				} else {
					data = append(data, "1")
				}

				// sends received
				sort.Strings(p.MercenariesReceivedPerWave[i])
				data = append(data, strings.Join(p.MercenariesReceivedPerWave[i], ","))

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

				for _, k := range keys {
					unit, ok := m[k]
					if !ok {
						data = append(data, "")
					} else {
						data = append(data, unit)
					}
				}

				err := w.Write(data)
				if err != nil || w.Error() != nil {
					panic(err)
				}
			}
		}
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