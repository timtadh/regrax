package stats

import (
	"log"
	"math"
	"math/rand"
)


func Srange(size int) []int {
	sample := make([]int, 0, size)
	for i := 0; i < size; i++ {
		sample = append(sample, i)
	}
	return sample
}

func Sample(size, populationSize int) (sample []int) {
	if size > populationSize {
		return Srange(populationSize)
	}
	pop := func(items []int) ([]int, int) {
		i := rand.Intn(len(items))
		item := items[i]
		copy(items[i:], items[i+1:])
		return items[:len(items)-1], item
	}
	items := Srange(populationSize)
	sample = make([]int, 0, size)
	for i := 0; i < size; i++ {
		var item int
		items, item = pop(items)
		sample = append(sample, item)
	}
	return sample
}

func ReplacingSample(size, populationSize int) (sample []int) {
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

func Round(val float64, places int ) (newVal float64) {
	var round float64
	roundOn := .5
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	_div := math.Copysign(div, val)
	_roundOn := math.Copysign(roundOn, val)
	if _div >= _roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}

func RandomPermutation(size int) (perm []int) {
	return Sample(size, size)
}

func Permutations(size int, do func(perm []int) (dobreak bool)) {
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
	if do(cur) {
		return
	}
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
				if do(cur) {
					return
				}
				break
			}
		}
		if i < 0 {
			break
		}
	}
}
