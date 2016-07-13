package main

import (
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	stage           = flag.String("stage", "staging", "env[STAGE]")
	entryCd         = flag.String("entry_cd", "", "cbankEntry.entry_cd")
	entryCdListFile = flag.String("list", "", "entry_cdと変更後URLのリストのcsvファイル")
	changeURL       = flag.String("url", "", "after url")
	dryRun          = flag.Bool("dry-run", true, "dry run")
	entryList       [][]string
	concurrent      = flag.Bool("concurrent", false, "concurrent")
)

func main() {
	flag.Parse()
	if len(*entryCdListFile) != 0 {
		entryList = parseEntryList()
	}
	if len(*entryCd) == 0 && len(entryList) == 0 {
		fmt.Println("entry_cdを指定してください")
		os.Exit(1)
	}

	if len(entryList) == 0 {
		vlo := NewVideoLinkObject(*stage, *entryCd)
		err := vlo.Bucket2Struct()
		if err != nil {
			fmt.Println("エラーです", err)
			os.Exit(1)
		}
		err = vlo.GenerateAfterVideoLink(*changeURL)
		if err != nil {
			fmt.Println("エラーです", err)
		}

		changeCount := vlo.ChangeVL()
		if changeCount > 0 {
			vlo.Write(*dryRun)
		}
	} else {
		if *concurrent == true {
			writeLinkConcurrent()
		} else {
			writeLink()
		}
	}
}

func writeLinkConcurrent() {
	var wg sync.WaitGroup

	for _, ecd := range entryList {
		if len(ecd[0]) == 0 {
			continue
		}
		wg.Add(1)
		go func(line []string) {
			vlo := NewVideoLinkObject(*stage, line[0])
			err := vlo.Bucket2Struct()
			if err != nil {
				fmt.Println("エラーです1", err)
				wg.Done()
				return
			}
			err = vlo.GenerateAfterVideoLink(line[1])
			if err != nil {
				fmt.Println("エラーです３", err)
			}
			changeCount := vlo.ChangeVL()
			if changeCount > 0 {
				vlo.Write(*dryRun)
			}
			wg.Done()
		}(ecd)
	}
	wg.Wait()
}

func writeLink() {
	for _, ecd := range entryList {
		if len(ecd[0]) == 0 {
			continue
		}

		vlo := NewVideoLinkObject(*stage, ecd[0])
		err := vlo.Bucket2Struct()
		if err != nil {
			fmt.Println("エラーです1", err)
			return
		}
		err = vlo.GenerateAfterVideoLink(ecd[1])
		if err != nil {
			fmt.Println("エラーです３", err)
		}
		changeCount := vlo.ChangeVL()
		if changeCount > 0 {
			vlo.Write(*dryRun)
		}
	}
}

func parseEntryList() [][]string {
	file, err := os.Open(*entryCdListFile)
	if err != nil {
		fmt.Println("リストファイルを取得できません：", err)
		return nil
	}
	defer file.Close()
	converter := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	scanner := bufio.NewScanner(converter)
	scanner.Split(CustomScan)
	header := 0

	var entryList [][]string
	for scanner.Scan() {
		if header == 0 {
			header = header + 1
			continue
		}
		line := strings.Split(scanner.Text(), ",")
		entryList = append(entryList, line)
	}
	return entryList
}

// 改行コードCR用
func CustomScan(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	var i int
	if i = bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	if i = bytes.IndexByte(data, '\r'); i >= 0 {
		// ここを追加した。（CR があったら、そこまでのデータを返そう）
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
