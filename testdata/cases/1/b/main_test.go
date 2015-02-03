package main

//go:generate tab

var ttDummyFunction = []struct {
	// inputs
	a, b int
	// ouputs
	c, d, e int
	f       float64
	err     error
}{
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil}, // and on and on ...
}

// TestTTDummyFunction is an automatically generated table driven test for the
// function DummyFunction using the tests defined in ttDummyFunction.
func TestTTDummyFunction(t *testing.T) {
	for i, tt := range ttDummyFunction {
		c, d, e, f, err := DummyFunction(tt.a, tt.b)
		if tt.c != c {
			t.Errorf("%d : expected %v, got %v", i, tt.c, c)
		}
		if tt.d != d {
			t.Errorf("%d : expected %v, got %v", i, tt.d, d)
		}
		if tt.e != e {
			t.Errorf("%d : expected %v, got %v", i, tt.e, e)
		}
		if tt.f != f {
			t.Errorf("%d : expected %v, got %v", i, tt.f, f)
		}
		if tt.err != err {
			t.Errorf("%d : expected %v, got %v", i, tt.err, err)
		}
	}
}
