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
	for _, tt := range ttDummyFunction {
		c, d, e, f, err := DummyFunction(tt.a, tt.b)
		if tt.c != c {
			t.Errorf("expected %v, got %v\n", tt.c, c)
		}
		if tt.d != d {
			t.Errorf("expected %v, got %v\n", tt.d, d)
		}
		if tt.e != e {
			t.Errorf("expected %v, got %v\n", tt.e, e)
		}
		if tt.f != f {
			t.Errorf("expected %v, got %v\n", tt.f, f)
		}
		if tt.err != err {
			t.Errorf("expected %v, got %v\n", tt.err, err)
		}
	}
}
