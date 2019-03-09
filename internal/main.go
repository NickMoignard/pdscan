package internal

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
)

func Main(urlStr string, showData bool, showAll bool, limit int) {
  f, err := os.Create("cpu.prof")
  if err != nil {
      log.Fatal("could not create CPU profile: ", err)
  }
  defer f.Close()
  if err := pprof.StartCPUProfile(f); err != nil {
      log.Fatal("could not start CPU profile: ", err)
  }
  defer pprof.StopCPUProfile()

	matchList := []ruleMatch{}

	if strings.HasPrefix(urlStr, "file://") || strings.HasPrefix(urlStr, "s3://") {
		files := findFiles(urlStr)

		if len(files) > 0 {
			fmt.Println(fmt.Sprintf("Found %s to scan...\n", pluralize(len(files), "file")))

			for _, file := range files {
				// fmt.Println("Scanning " + file + "...\n")
				// TODO use streaming instead
				// TODO process in parallel
				values := readLines(file)
				matchList = findMatches(file, values, true)
				printMatchList(matchList, showData, showAll, "line")
			}
		} else {
			fmt.Println("Found no files to scan")
			return
		}
	} else {
		var adapter Adapter = &SqlAdapter{}
		adapter.Init(urlStr)

		tables := adapter.FetchTables()

		if len(tables) > 0 {
			fmt.Println(fmt.Sprintf("Found %s to scan, sampling %d rows from each...\n", pluralize(len(tables), "table"), limit))

			for _, table := range tables {
				columnNames, columnValues := adapter.FetchTableData(table, limit)
				tableMatchList := checkTableData(table, columnNames, columnValues)
				matchList = append(matchList, tableMatchList...)
				printMatchList(tableMatchList, showData, showAll, "row")
			}
		} else {
			fmt.Println("Found no tables to scan")
			return
		}
	}

	if len(matchList) > 0 {
		if showData {
			fmt.Println("Showing 50 unique values from each")
		} else {
			fmt.Println("\nUse --show-data to view data")
		}

		if !showAll {
			showLowConfidenceMatchHelp(matchList)
		}
	} else {
		fmt.Println("No sensitive data found")
	}
}
