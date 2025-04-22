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

const chars = "abcdefghijklmnopqrstuvwxyz"

type result struct {
	chosen string
	score  int
	count  int // testing number of iterations
}

func routine(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc, ch chan<- result, goal string, id int) {
	defer wg.Done()
	fmt.Printf("Running routine %d\n", id)
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	generated := make([]byte, len(goal))

	res := result{
		chosen: "",
		score:  0,
		count:  0,
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
					generated[i] = chars[seed.Intn(len(chars))]
					if goal[i] == generated[i] {
						tmpScore++
					} else {
						break
					}
				}
				if tmpScore > res.score {
					res.score = tmpScore
					res.chosen = string(generated[0:tmpScore])
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

	fmt.Printf("Best score achieved was %d with '%s'. A total of %d iterations were achieved\n", results[0].score, results[0].chosen, finalCount)
}
