package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type Stats struct {
	Url         string
	Connections int
	Threads     int
	AvgDuration time.Duration
	Duration    time.Duration
	Sum         time.Duration
	Times       []int
	Transferred int64
	Resp200     int64
	Resp300     int64
	Resp400     int64
	Resp500     int64
}

func CalcStats(responseChannel chan response, startTime time.Time) []byte {

	stats := &Stats{
		Url:         target,
		Connections: *numConnections,
		Threads:     *numThreads,
		Times:       make([]int, 0),
	}

	for r := range responseChannel {
		stats.Sum += r.duration
		stats.Times = append(stats.Times, int(r.duration))
		stats.Transferred += r.size

		switch {
		case r.code < 300:
			stats.Resp200++
		case r.code < 400:
			stats.Resp300++
		case r.code < 500:
			stats.Resp400++
		case r.code < 600:
			stats.Resp500++
		}
	}
	// done ranging so requests are done
	stats.Duration = time.Now().Sub(startTime)
	if len(stats.Times) > 0 {
		stats.AvgDuration = time.Duration(stats.Duration.Nanoseconds() / int64(len(stats.Times)))
	}
	PrintStats(stats)
	b, err := json.Marshal(&stats)
	if err != nil {
		fmt.Println(err)
	}
	return b
}

func CalcDistStats(distChan chan string) {
	if len(distChan) == 0 {
		return
	}
	allStats := &Stats{
		Url:         target,
		Connections: *numConnections,
		Threads:     *numThreads,
	}
	statCount := len(distChan)
	for res := range distChan {
		var stats Stats
		err := json.Unmarshal([]byte(res), &stats)
		if err != nil {
			fmt.Println(err)
		}
		allStats.Duration += stats.Duration
		allStats.Times = append(allStats.Times, stats.Times...)
		allStats.Resp200 += stats.Resp200
		allStats.Resp300 += stats.Resp300
		allStats.Resp400 += stats.Resp400
		allStats.Resp500 += stats.Resp500
		if len(distChan) == 0 {
			break
		}
	}
	allStats.AvgDuration = time.Duration(allStats.Duration.Nanoseconds() / int64(statCount))
	PrintStats(allStats)
}

func PrintStats(allStats *Stats) {
	sort.Ints(allStats.Times)
	total := float64(len(allStats.Times))
	totalInt := int64(total)
	fmt.Println("==========================BENCHMARK==========================")
	fmt.Printf("URL:\t\t\t\t%s\n\n", allStats.Url)
	fmt.Printf("Used Connections:\t\t%d\n", allStats.Connections)
	fmt.Printf("Used Threads:\t\t\t%d\n", allStats.Threads)
	fmt.Printf("Total number of calls:\t\t%d\n", totalInt)
	fmt.Println("")
	fmt.Println("============================TIMES============================")
	fmt.Printf("Total time passed:\t\t%.2fs\n", allStats.Duration.Seconds())
	fmt.Printf("Avg time per request:\t\t%.2fms\n", float64(allStats.Sum)/total/1e6)
	fmt.Printf("Requests per second:\t\t%.2f\n", total/allStats.Duration.Seconds())
	fmt.Printf("Median time per request:\t%.2fms\n", float64(allStats.Times[(totalInt-1)/2]/1e6))
	fmt.Printf("99th percentile time:\t\t%.2fms\n", float64(allStats.Times[(totalInt/100*99)])/1e6)
	fmt.Printf("Slowest time for request:\t%.2fms\n", float64(allStats.Times[totalInt-1]/1e6))
	mb := float64(allStats.Transferred) / 1024 / 1024
	fmt.Printf("Total Body bytes transferred:\t%.2fMB\n", mb)
	fmt.Printf("Transfer rate:\t\t\t%.2fMB/s\n", mb/float64(allStats.Duration.Seconds()))
	fmt.Println("")
	fmt.Println("==========================RESPONSES==========================")
	fmt.Printf("20X allStats.Responses:\t\t%d\t(%d%%)\n", allStats.Resp200, allStats.Resp200/totalInt*100)
	fmt.Printf("30X allStats.Responses:\t\t%d\t(%d%%)\n", allStats.Resp300, allStats.Resp300/totalInt*100)
	fmt.Printf("40X allStats.Responses:\t\t%d\t(%d%%)\n", allStats.Resp400, allStats.Resp400/totalInt*100)
	fmt.Printf("50X allStats.Responses:\t\t%d\t(%d%%)\n", allStats.Resp500, allStats.Resp500/totalInt*100)
}
