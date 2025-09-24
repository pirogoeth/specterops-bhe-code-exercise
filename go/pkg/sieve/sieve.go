package sieve

import (
	"math"
	"os"

	"github.com/bits-and-blooms/bitset"
	"github.com/davecgh/go-spew/spew"
)

type Sieve interface {
	NthPrime(n int64) int64
}

func Debug(items ...any) {
	if os.Getenv("DEBUG_ME") != "" {
		spew.Dump(items...)
	}
}

// NewSieve creates an instance of the sieve impl without any caching
func NewSieve() Sieve {
	return &sieveImpl{
		marked:            nil,
		largestUpperBound: 0,
	}
}

type sieveImpl struct {
	// marked is a running bitset of the numbers that have been cleared
	// from the prime sieve
	marked *bitset.BitSet
	// largestUpperBound tracks the largest size we have marked, so on larger values of N,
	// we can make use of the previously marked sets as an optimization and only fill in as needed
	largestUpperBound int64
}

// algorithm Sieve of Eratosthenes is
//     input: an integer n > 1.
//     output: all prime numbers from 2 through n.
//
//     let A be an array of Boolean values, indexed by integers 2 to n,
//     initially all set to true.
//
//     for i = 2, 3, 4, ..., not exceeding √n do
//         if A[i] is true
//             for j = i^2, i^2+i, i^2+2i, i^2+3i, ..., not exceeding n do
//                 set A[j] := false
//
//     return all i such that A[i] is true.
//
//
// But there are some neat optimizations we can do here!
//
// First thing is that we can use Rosser's theorem to find the (_rough_) upper bound of a prime
// at position _n_
//
// Only ever check odds, since they are the only "flavor" of number that can be prime, 2 aside
//
// Retain the marks between returns instead of recomputing each NthPrime run. This gives us a modest
// speedup with the complication that we need to keep track of our previous upper bounds
// and calculate a rough tail to pick up marking from

func (s *sieveImpl) NthPrime(num int64) int64 {
	fNum := float64(num)
	upperBound := int64(6)
	if num > 1 {
		// Use Rosser's theorem to rough out upper bounds for the Nth prime
		upperBound = int64(math.Ceil(fNum*(math.Log(fNum)+math.Log(math.Log(fNum))))*1.05 + 10)
	}

	if s.largestUpperBound == 0 {
		s.marked = bitset.New(uint(upperBound) + 1)
		s.marked.SetAll()
		// Only mark odds (and 2)
		s.marked.Set(2)
		for idx := uint(3); idx < s.marked.Len(); idx += 2 {
			s.marked.Set(idx)
		}
	} else if upperBound > s.largestUpperBound && s.largestUpperBound != 0 {
		// When we have previously marked out numbers, we want to avoid the extra effort
		// by only marking the delta between the previous upper bound and the new upper bound
		markStart := uint(s.largestUpperBound)
		if markStart%2 == 0 {
			markStart += 1
		}

		for idx := markStart; idx <= uint(upperBound); idx += 2 {
			s.marked.Set(idx)
		}
	}

	// Only sieve up to the square root of the new upper bound.
	// Anything larger than that will be > the upper bound
	basePrimeUpperLimit := int64(math.Ceil(math.Sqrt(float64(upperBound))))
	for i := int64(2); i < basePrimeUpperLimit; i++ {
		if s.marked.Test(uint(i)) {
			// Pick up from the tail end of the previous base primes. For smaller primes, the previous
			// largest bound _COULD_ put you before i*i, which doesn't need to be remarked.
			multiplesStart := int64(math.Max(float64(i*i), math.Ceil(float64(s.largestUpperBound/i)*float64(i))))
			for j := multiplesStart; j < upperBound; j += i {
				s.marked.Clear(uint(j))
			}
		}
	}

	s.largestUpperBound = upperBound
	primes := s.generatePrimes()

	Debug(
		"marked", s.marked,
		"num", num,
		"largestUpperBound", s.largestUpperBound,
		"upperBound", upperBound,
		"primes", primes,
	)

	return primes[num]
}

func (s *sieveImpl) generatePrimes() []int64 {
	primes := []int64{}

	for i := int64(2); i < s.largestUpperBound; i++ {
		if s.marked.Test(uint(i)) {
			primes = append(primes, i)
		}
	}

	return primes
}
