package math

import "testing"

func TestRandomInRange(t *testing.T) {
    min, max := 1, 5
    for i := 0; i < 100; i++ {
        v := Random(min, max)
        if v < min || v >= max {
            t.Fatalf("Random(%d,%d) = %d out of range", min, max, v)
        }
    }
}

func TestRandomNegativeRange(t *testing.T) {
    min, max := -5, 0
    for i := 0; i < 50; i++ {
        v := Random(min, max)
        if v < min || v >= max {
            t.Fatalf("Random(%d,%d) = %d out of range", min, max, v)
        }
    }
}

func TestRandomSingleValueRange(t *testing.T) {
    // when max = min+1 the only valid result is min
    min, max := 10, 11
    for i := 0; i < 10; i++ {
        v := Random(min, max)
        if v != min {
            t.Fatalf("Random(%d,%d) = %d, want %d", min, max, v, min)
        }
    }
}

func TestRandomPanicsOnInvalidRange(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Fatalf("expected panic when min >= max")
        }
    }()
    // min == max should cause runtime panic from Intn(0)
    Random(5, 5)
}
