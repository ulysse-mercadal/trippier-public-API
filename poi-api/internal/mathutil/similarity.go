package mathutil

// JaroWinkler returns a string similarity score in [0, 1] using the Jaro-Winkler metric.
// A score of 1.0 means identical strings; 0.0 means completely dissimilar.
func JaroWinkler(s1, s2 string) float64 {
	jaro := jaroSimilarity(s1, s2)
	prefixLen := 0
	limit := minInt(4, minInt(len(s1), len(s2)))
	for i := 0; i < limit; i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}
	return jaro + float64(prefixLen)*0.1*(1-jaro)
}

// jaroSimilarity computes the base Jaro similarity score in [0, 1] before the Winkler prefix bonus is applied.
func jaroSimilarity(s1, s2 string) float64 {
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}
	matchDist := maxInt(len(s1), len(s2))/2 - 1
	if matchDist < 0 {
		matchDist = 0
	}
	s1Matches := make([]bool, len(s1))
	s2Matches := make([]bool, len(s2))
	matches := 0
	transpositions := 0

	for i := range s1 {
		start := maxInt(0, i-matchDist)
		end := minInt(i+matchDist+1, len(s2))
		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0.0
	}
	k := 0
	for i := range s1 {
		if !s1Matches[i] {
			continue
		}
		for k < len(s2) && !s2Matches[k] {
			k++
		}
		if k < len(s2) && s1[i] != s2[k] {
			transpositions++
		}
		k++
	}
	m := float64(matches)
	return (m/float64(len(s1)) + m/float64(len(s2)) + (m-float64(transpositions)/2)/m) / 3
}

// minInt returns the smaller of a and b.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the larger of a and b.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
