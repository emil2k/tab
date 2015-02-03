package main

//go:generate tab

var ttDummyFunction = []struct {
	// inputs
	a, b int
	// outputs
	c, d, e int
	f       float64
	err     error
}{
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil}, // and on and on ...
}
