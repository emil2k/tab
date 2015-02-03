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

// TestTTDummyFunction is an automatically generated table driven test for the
// function DummyFunction using the tests defined in ttDummyFunction.
func TestTTDummyFunction(t *testing.T) {
	for i, tt := range ttDummyFunction {
		c, d, e, f, err := DummyFunction(tt.a, tt.b)
		if c != tt.c {
			t.Errorf("%d : c : got %v, expected %v", i, c, tt.c)
		}
		if d != tt.d {
			t.Errorf("%d : d : got %v, expected %v", i, d, tt.d)
		}
		if e != tt.e {
			t.Errorf("%d : e : got %v, expected %v", i, e, tt.e)
		}
		if f != tt.f {
			t.Errorf("%d : f : got %v, expected %v", i, f, tt.f)
		}
		if err != tt.err {
			t.Errorf("%d : err : got %v, expected %v", i, err, tt.err)
		}
	}
}
