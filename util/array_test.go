package util

import "testing"

func TestContain(t *testing.T) {
    tests := []struct{
        name string
        obj interface{}
        target interface{}
        want bool
    }{
        {"string in slice", "a", []string{"a","b"}, true},
        {"int in slice", 2, []int{1,2,3}, true},
        {"string not in slice", "x", []string{"a","b"}, false},
        {"array contains", "z", [3]string{"x","y","z"}, true},
        {"map contains key", "key", map[string]int{"key":1}, true},
        {"unsupported target type", 1, 123, false},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Contain(tc.obj, tc.target)
            if tc.want {
                if !got || err != nil {
                    t.Fatalf("Contain(%v, %v) = (%v, %v), want true,nil", tc.obj, tc.target, got, err)
                }
            } else {
                if got || err == nil {
                    t.Fatalf("Contain(%v, %v) = (%v, %v), want false,err", tc.obj, tc.target, got, err)
                }
            }
        })
    }
}

func TestURIContain(t *testing.T) {
    tests := []struct{
        name string
        path string
        list []string
        want bool
    }{
        {"contains segment", "/api/noauth/login", []string{"noauth","static"}, true},
        {"not contained", "/api/auth/login", []string{"noauth","static"}, false},
        {"empty list", "/any/path", []string{}, false},
        {"substring mid", "/assets/static/js/app.js", []string{"static/js"}, true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := URIContain(tc.path, tc.list)
            if tc.want {
                if !got || err != nil {
                    t.Fatalf("URIContain(%q, %v) = (%v, %v), want true,nil", tc.path, tc.list, got, err)
                }
            } else {
                if got || err == nil {
                    t.Fatalf("URIContain(%q, %v) = (%v, %v), want false,err", tc.path, tc.list, got, err)
                }
            }
        })
    }
}
