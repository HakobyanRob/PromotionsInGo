package main

import (
	"container/list"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

var mu sync.Mutex

/*func main() {
	f, _ := os.Open("promotions.csv")
	f1, _ := os.Open("promotions.csv")
	defer f.Close()
	defer f1.Close()

	ts := time.Now()
	basicRead(f)
	te := time.Now().Sub(ts)

	ts1 := time.Now()
	basicReadAll()
	//readFromCSVConc()

	te1 := time.Now().Sub(ts1)

	// Read and Set to a map
	fmt.Println("\nEND Basic: ", te)
	fmt.Println("END Concu: ", te1)
}*/

func readFromCSVConc() []Promotion {
	var wg sync.WaitGroup
	rows := list.New()
	var promotions []Promotion
	wg.Add(1)
	fillQueue(rows, &wg)
	wg.Add(1)
	parseQueue(rows, promotions, &wg)
	wg.Wait()
	return promotions
}

func parseQueue(rows *list.List, promotions []Promotion, wg *sync.WaitGroup) {
	defer wg.Done()
	for rows.Len() > 0 {
		e := rows.Front() // First element
		rows.Remove(e)    // Dequeue

		strings := e.Value.([]string)
		price, _ := strconv.ParseFloat(strings[1], 64)
		p := Promotion{
			strings[0], price, strings[2],
		}
		promotions = append(promotions, p)
	}
}

func fillQueue(rows *list.List, wg *sync.WaitGroup) {
	defer wg.Done()
	filePath := "promotions.csv"
	open, err2 := os.Open(filePath)
	f, err := open, err2
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	for {
		rStr, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			break
		}
		rows.PushBack(rStr)
	}
}
