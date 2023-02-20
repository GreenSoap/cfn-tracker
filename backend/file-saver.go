package backend

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func ResetSaveData() {
	SaveMatchHistory(MatchHistory{
		CFN:          ``,
		LP:           0,
		LPGain:       0,
		Wins:         0,
		Losses:       0,
		TotalWins:    0,
		TotalLosses:  0,
		TotalMatches: 0,
		WinRate:      0,
		IsWin:        false,
	})
}

func GetLastSavedMatchHistory() (MatchHistory, bool) {
	var result MatchHistory

	pastResultsFile, err := os.ReadFile(`results/results.json`)
	if err != nil {
		return MatchHistory{}, false
	}

	_ = json.Unmarshal(pastResultsFile, &result)
	if err != nil {
		fmt.Println(`Could not unmarshal match history`, err)
		return MatchHistory{}, false
	}

	return result, true
}

func SaveMatchHistory(matchHistory MatchHistory) {
	SaveTextToFile(`results`, `wins.txt`, strconv.Itoa(matchHistory.Wins))
	SaveTextToFile(`results`, `losses.txt`, strconv.Itoa(matchHistory.Losses))
	SaveTextToFile(`results`, `win-rate.txt`, strconv.Itoa(matchHistory.WinRate)+`%`)
	SaveTextToFile(`results`, `win-streak.txt`, strconv.Itoa(matchHistory.WinStreak))
	SaveTextToFile(`results`, `lp.txt`, strconv.Itoa(matchHistory.LP))
	gain := strconv.Itoa(matchHistory.LPGain)
	if matchHistory.LPGain > 0 {
		gain = `+` + gain
	}
	SaveTextToFile(`results`, `lp-gain.txt`, gain)

	// Do not save match result if there is no opponent
	if matchHistory.Opponent == `` {
		return
	}
	mhMarshalled, err := json.Marshal(&matchHistory)

	if err != nil {
		return
	}

	// Save current results
	SaveTextToFile(`results`, `results.json`, string(mhMarshalled))

	// Now save current results to the entire log

	var arr []MatchHistory

	pastMatches, err := os.ReadFile(`results/` + matchHistory.CFN + `-log.json`)
	if err != nil {
		// No past matches
		str := "[" + string(mhMarshalled) + "]"
		SaveTextToFile(`results`, matchHistory.CFN+`-log.json`, str)
		return
	}

	err = json.Unmarshal(pastMatches, &arr)
	if err != nil {
		fmt.Println(`Could not unmarshal match history`, err)
		return
	}

	newArr := append(arr, matchHistory)
	newArrMarshalled, err := json.Marshal(&newArr)
	if err != nil {
		fmt.Println(`Error marshalling match history`)
		return
	}
	fmt.Println(string(newArrMarshalled))
	SaveTextToFile(`results`, matchHistory.CFN+`-log.json`, string(newArrMarshalled))
}

func SaveTextToFile(directory string, fileName string, text string) {
	var file *os.File
	var err error

	if directory != `` {
		err = os.Mkdir(`results`, os.FileMode(0755))
		file, err = os.Create(directory + `/` + fileName)
	} else {
		file, err = os.Create(fileName)
	}

	_, err = file.WriteString(text)
	defer file.Close()
	if err != nil {
		fmt.Println(`Issue writing to file`, fileName)
		LogError(SaveError)
	}
}

func (a *App) ExportLogToCSV(cfn string) {
	var matchHistories []MatchHistory
	pastMatches, _ := os.ReadFile(`results/` + cfn + `-log.json`)
	_ = json.Unmarshal(pastMatches, &matchHistories)
	csvFile, _ := os.Create(`results/` + cfn + `.csv`)
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)

	// Header
	var header []string
	header = append(header, "Date")
	header = append(header, "Time")
	header = append(header, "Opponent")
	header = append(header, "Opponent Character")
	header = append(header, "Opponent LP")
	header = append(header, "Result")

	writer.Write(header)

	for _, obj := range matchHistories {
		var record []string
		record = append(record, obj.Date)
		record = append(record, obj.TimeStamp)
		record = append(record, obj.Opponent)
		record = append(record, obj.OpponentCharacter)
		record = append(record, obj.OpponentLP)

		if obj.IsWin == true {
			record = append(record, `W`)
		} else if obj.IsWin == false {
			record = append(record, `L`)
		}

		writer.Write(record)
		record = nil
	}

	writer.Flush()
}
