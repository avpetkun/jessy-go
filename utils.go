package jessy

func slicesGetFrame[S ~[]E, E any](s S, n int) (newS, frame S) {
	s0 := len(s)
	s1 := s0 + n
	if s1 > cap(s) {
		s = append(s, make([]E, n)...)
	}
	newS = s[:s1]
	frame = s[s0:s1]
	return
}
