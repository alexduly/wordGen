package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

var charSlice = []byte("abcdefghijklmnopqrstuvwxyz")

type result struct {
	score int
	count int // testing number of iterations
}

func routine(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc, ch chan<- result, goal string, id int) {
	defer wg.Done()
	seedVal := time.Now().UnixNano() ^ int64(id)*1_000_000_007
	seed := rand.New(rand.NewSource(seedVal))
	generated := make([]byte, len(goal))

	res := result{
		score: 0,
		count: 0,
	}
	for {
		select {
		case <-ctx.Done():
			ch <- res
			return
		default:
			{
				res.count++
				tmpScore := 0
				for i := 0; i < len(goal); i++ {
					generated[i] = charSlice[seed.Intn(26)]
					if goal[i] == generated[i] {
						tmpScore++
					} else {
						break
					}
				}
				if tmpScore > res.score {
					res.score = tmpScore
					if res.score == len(goal) { // cancel early if a full score is achieved...
						deadline, _ := ctx.Deadline()
						timeleft := time.Until(deadline).Seconds()
						fmt.Printf("Routine %d has achieved max score, cancelling all and returning. Process finished: %.2fs early\n", id, timeleft)
						ch <- res
						cancel()
					}
				}
			}
		}
	}
}

func main() {
	maxProcs := runtime.GOMAXPROCS(0) + 2 // Gets max num of cores we can use, add 2 as some will get blocked occasionally
	var wg sync.WaitGroup
	wg.Add(maxProcs)
	const goal = "shakespeare"
	ch := make(chan result, maxProcs)
	fmt.Printf("Starting at %s with %d processes\n", time.Now().String(), maxProcs)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for i := 0; i < maxProcs; i++ {
		go routine(ctx, &wg, cancel, ch, goal, i)
	}

	wg.Wait()

	close(ch)
	finalCount := 0
	var results []result
	for msg := range ch {
		results = append(results, msg)
		finalCount += msg.count
	}

	sort.Slice(results, func(i int, j int) bool {
		return results[i].score > results[j].score
	})

	fmt.Printf("Best score achieved was %d with '%s'. A total of %d iterations were achieved\n", results[0].score, goal[0:results[0].score], finalCount)
}
