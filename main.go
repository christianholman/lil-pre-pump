package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/mitchellh/colorstring"
	bittrex "github.com/toorop/go-bittrex"
)

type Configuration struct {
	APIKey     string  `json:"api_key"`
	APISecret  string  `json:"api_secret"`
	MinChange  float64 `json:"min_change"`
	MinVolume  float64 `json:"min_volume"`
	UpdateTime float64 `json:"update_time"`
	MaxVolume  float64 `json:"max_volume"`
}

func main() {

	// LOAD CONFIG
	file, _ := os.Open("botsettings.json")
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}

	bittrex := bittrex.New(config.APIKey, config.APISecret)

	startSumVolume := make(map[string]float64)
	startSumPrice := make(map[string]float64)
	trackedCoins := make(map[string]bool)

	ShowIntro(config.MinChange, config.MinVolume, config.MaxVolume)

	sum, err := bittrex.GetMarketSummaries()

	if err != nil {
		panic(err)
	}

	for i := range sum {
		startSumVolume[sum[i].MarketName] = sum[i].BaseVolume
		startSumPrice[sum[i].MarketName] = sum[i].Last
	}

	for true {
		summaries, err := bittrex.GetMarketSummaries()

		found := 0

		if err != nil {
			panic(err)
		}

		for i := range summaries {
			change := PercentageChange(startSumVolume[summaries[i].MarketName], summaries[i].BaseVolume)
			priceChange := PercentageChange(startSumPrice[summaries[i].MarketName], summaries[i].Last)
			volIncrease := summaries[i].BaseVolume - startSumVolume[summaries[i].MarketName]
			volChangePercent := PercentageChange(startSumVolume[summaries[i].MarketName], summaries[i].BaseVolume)
			coinVolume := startSumVolume[summaries[i].MarketName]

			if change > config.MinChange && volIncrease > config.MinVolume && strings.HasPrefix(summaries[i].MarketName, "BTC") && priceChange > 0 && config.MaxVolume > coinVolume {

				found++

				if _, ok := trackedCoins[summaries[i].MarketName]; !ok {
					trackedCoins[summaries[i].MarketName] = true
				} else {
					trackedCoins[summaries[i].MarketName] = false
				}

				percColor := "[green]"

				str := fmt.Sprintf("%v%v [yellow]%v [white]%v[green](%v%v) [white]%v [green](+%v / %v%v%v[green]) ",
					ifNew(trackedCoins, summaries[i].MarketName),
					time.Now().Format("15:04:05"),
					summaries[i].MarketName,
					summaries[i].Last,
					toFixed(priceChange, 2),
					"%",
					toFixed(summaries[i].BaseVolume, 9),
					toFixed(volIncrease, 2),
					percColor,
					toFixed(volChangePercent, 2),
					"%",
				)

				colorstring.Fprintln(ansi.NewAnsiStdout(), str)
			} else {
				delete(trackedCoins, summaries[i].MarketName)
			}
		}

		if found != 0 {
			fmt.Print("\n")
		}

		time.Sleep(time.Duration(config.UpdateTime) * time.Second)
	}
}

func ShowIntro(minChange, minVolume, maxVolume float64) {
	str := fmt.Sprintf(`
[green]================================================================
[green]// [yellow]LIL PRE-PUMP [red]PUMP [yellow]TOOL - v1.0 - November 29, 2017 - Author: [red]Luxio 
[green]================================================================
[green]// Minimum Price Change: [yellow]%v%v					  
[green]// Minimum Volume Change: [yellow]%v BTC	
[green]// Maximum Market Volume: [yellow]%v BTC					  	
[green]================================================================
`, minChange, "%", minVolume, maxVolume,
	)

	colorstring.Fprintln(ansi.NewAnsiStdout(), str)
}

func PercentageChange(old, new float64) (delta float64) {
	diff := float64(new - old)
	delta = (diff / float64(old)) * 100
	return
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func ifNew(trackedCoins map[string]bool, market string) string {
	if trackedCoins[market] {
		return "[green]"
	}
	return ""
}
