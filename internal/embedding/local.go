package embedding

import (
	"hash/fnv"
	"math"
	"regexp"
	"strings"
)

const Dimensions = 64

var tokenPattern = regexp.MustCompile(`[\pL\pN][\pL\pN_-]*`)

type Vector [Dimensions]float64

func Embed(text string) Vector {
	var vector Vector
	for _, token := range tokenPattern.FindAllString(strings.ToLower(text), -1) {
		if len(token) < 2 {
			continue
		}
		h := fnv.New64a()
		_, _ = h.Write([]byte(token))
		value := h.Sum64()
		idx := int(value % Dimensions)
		sign := 1.0
		if value&(1<<63) != 0 {
			sign = -1
		}
		vector[idx] += sign
	}
	return Normalize(vector)
}

func Normalize(vector Vector) Vector {
	norm := 0.0
	for _, value := range vector {
		norm += value * value
	}
	if norm == 0 {
		return vector
	}
	norm = math.Sqrt(norm)
	for i := range vector {
		vector[i] /= norm
	}
	return vector
}

func Cosine(left Vector, right Vector) float64 {
	score := 0.0
	for i := range left {
		score += left[i] * right[i]
	}
	return score
}
