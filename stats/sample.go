package stats

import (
	"encoding/binary"
	"log"
	"math/rand"
	"os"
)


func init() {
	if urandom, err := os.Open("/dev/urandom"); err != nil {
		panic(err)
	} else {
		seed := make([]byte, 8)
		if _, err := urandom.Read(seed); err == nil {
			rand.Seed(int64(binary.BigEndian.Uint64(seed)))
		}
		urandom.Close()
	}
}

func Srange(size int) []int {
	sample := make([]int, 0, size)
	for i := 0; i < size; i++ {
		sample = append(sample, i)
	}
	return sample
}

func Sample(size, populationSize int) (sample []int) {
	if size >= populationSize {
		return Srange(populationSize)
	}
	in := func(x int, items []int) bool {
		for _, y := range items {
			if x == y {
				return true
			}
		}
		return false
	}
	sample = make([]int, 0, size)
	for i := 0; i < size; i++ {
		j := rand.Intn(populationSize)
		for in(j, sample) {
			j = rand.Intn(populationSize) 
		}
		sample = append(sample, j)
	}
	return sample
}

func ReplacingSample(size, populationSize int) (sample []int) {
	if size >= populationSize {
		return Srange(populationSize)
	}
	sample = make([]int, 0, size)
	for i := 0; i < size; i++ {
		j := rand.Intn(populationSize)
		sample = append(sample, j)
	}
	return sample
}

func WeightedSample(prs []float64) int {
	total := Sum(prs)
	i := 0
	x := total * (1 - rand.Float64())
	for x > prs[i] {
		x -= prs[i]
		i += 1
	}
	return i
}

func Min(items []int, f func(item int) float64) (arg int, min float64) {
	arg = -1
	for _, i := range items {
		d := f(i)
		if d < min || arg < 0 {
			min = d
			arg = i
		}
	}
	return arg, min
}

func Max(items []int, f func(item int) float64) (arg int, max float64) {
	arg = -1
	for _, i := range items {
		d := f(i)
		if d > max || arg < 0 {
			max = d
			arg = i
		}
	}
	if arg < 0 {
		log.Panic("arg < 0")
	}
	return arg, max
}

func Sum(list []float64) float64 {
	var sum float64
	for _, item := range list {
		sum += item
	}
	return sum
}

func Permutations(size int) (results [][]int) {
	slice := make([]int, size)
	indices := make([]int, size)
	cycles := make([]int, size)
	cur := make([]int, size)
	prev := cur
	for i := 0; i < size; i++ {
		slice[i] = i
		indices[i] = i
		cycles[i] = size - i
		cur[i] = i
	}
	results = append(results, cur)
	for {
		prev = cur
		cur = make([]int, size)
		copy(cur, prev)
		i := size - 1
		for ; i >= 0; i -= 1 {
			cycles[i] -= 1
			if cycles[i] == 0 {
				index := indices[i]
				for j := i; j < size-1; j += 1 {
					indices[j] = indices[j+1]
				}
				indices[size-1] = index
				cycles[i] = size - i
			} else {
				j := cycles[i]
				indices[i], indices[size-j] = indices[size-j], indices[i]
				for k := i; k < size; k += 1 {
					cur[k] = slice[indices[k]]
				}
				results = append(results, cur)
				break
			}
		}
		if i < 0 {
			break
		}
	}
	return results
}

