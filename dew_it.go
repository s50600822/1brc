/*
Copyright 2023 The original authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func processChunk(ch <-chan string, wg *sync.WaitGroup, stationTemperatures chan<- map[string][]float64) {
	defer wg.Done()
	stationTemperaturesMap := make(map[string][]float64)

	for line := range ch {
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			continue
		}
		station := parts[0]
		temperature, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		stationTemperaturesMap[station] = append(stationTemperaturesMap[station], temperature)
	}

	stationTemperatures <- stationTemperaturesMap
}

func mergeResults(results ...map[string][]float64) map[string][]float64 {
	mergedResults := make(map[string][]float64)
	for _, result := range results {
		for station, temperatures := range result {
			mergedResults[station] = append(mergedResults[station], temperatures...)
		}
	}
	return mergedResults
}

func processTemperatures(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var chunks [][]string
	for scanner.Scan() {
		line := scanner.Text()
		chunks = append(chunks, []string{line})
	}

	// Number of concurrent workers
	numWorkers := 10
	chunkSize := (len(chunks) + numWorkers - 1) / numWorkers

	// Channel to receive temperature data from workers
	stationTemperaturesChan := make(chan map[string][]float64)

	// Start workers
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(chunks) {
			end = len(chunks)
		}
		go processChunk(createChunkChannel(chunks[start:end]), &wg, stationTemperaturesChan)
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(stationTemperaturesChan)
	}()

	// Collect results from workers
	var results []map[string][]float64
	for temp := range stationTemperaturesChan {
		results = append(results, temp)
	}

	mergedResults := mergeResults(results...)

	// Calculate min, mean, and max temperatures for each station
	resultsMap := make(map[string]string)
	for station, temperatures := range mergedResults {
		minTemp := round(min(temperatures), 1)
		meanTemp := round(sum(temperatures)/float64(len(temperatures)), 1)
		maxTemp := round(max(temperatures), 1)
		resultsMap[station] = fmt.Sprintf("%.1f/%.1f/%.1f", minTemp, meanTemp, maxTemp)
	}

	// Sort results alphabetically by station name
	var keys []string
	for key := range resultsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Print("{")
	for i, key := range keys {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s=%s", key, resultsMap[key])
	}
	fmt.Println("}")
}

func createChunkChannel(chunk [][]string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, line := range chunk {
			ch <- line[0]
		}
	}()
	return ch
}

func min(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	min := slice[0]
	for _, value := range slice {
		if value < min {
			min = value
		}
	}
	return min
}

func max(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	max := slice[0]
	for _, value := range slice {
		if value > max {
			max = value
		}
	}
	return max
}

func sum(slice []float64) float64 {
	total := 0.0
	for _, value := range slice {
		total += value
	}
	return total
}

func round(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	if intermed < 0.0 {
		intermed -= 0.5
	} else {
		intermed += 0.5
	}
	rounder = float64(int64(intermed))
	return rounder / pow
}

func main() {
	startTime := time.Now()
	defer func() {
		fmt.Printf("Elapsed time: %v\n", time.Since(startTime))
	}()

	processTemperatures("measurements.txt")
}
