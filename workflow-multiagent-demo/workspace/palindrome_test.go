package palindrome

import "testing"

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"a", true},
		{"aa", true},
		{"ab", false},
		{"A man, a plan, a canal: Panama", true},
		{"race a car", false},
		{"0P", false},
		{"12321", true},
		{"Able was I ere I saw Elba", true},
		{"No 'x' in Nixon", true},
	}
	for _, tt := range tests {
		got := IsPalindrome(tt.input)
		if got != tt.want {
			t.Errorf("IsPalindrome(%q) = %v; want %v", tt.input, got, tt.want)
		}
	}
}
