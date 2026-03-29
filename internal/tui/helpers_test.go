package tui

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 5, "hell…"},
		{"single char max", "hello", 1, "…"},
		{"zero max", "hello", 0, ""},
		{"negative max", "hello", -1, ""},
		{"empty string", "", 5, ""},
		{"max 2", "hello", 2, "h…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		w       int
		wantLen int
	}{
		{"shorter than width", "hi", 10, 10},
		{"exact width", "hello", 5, 5},
		{"longer than width", "hello world", 5, 11},
		{"empty string", "", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padRight(tt.s, tt.w)
			// padRight uses lipgloss.Width which may differ from len for styled strings
			// For plain strings they should be equivalent
			if len(got) != tt.wantLen {
				t.Errorf("padRight(%q, %d) has length %d, want %d", tt.s, tt.w, len(got), tt.wantLen)
			}
		})
	}
}

func TestLayoutHelpers(t *testing.T) {
	m := model{width: 100, height: 50}

	t.Run("listW", func(t *testing.T) {
		got := m.listW()
		want := 40 // 100*38/100=38, clamped up to listMinW=40
		if got != want {
			t.Errorf("listW() = %d, want %d", got, want)
		}
	})

	t.Run("detailW", func(t *testing.T) {
		got := m.detailW()
		want := 60 // 100 - 40
		if got != want {
			t.Errorf("detailW() = %d, want %d", got, want)
		}
	})

	t.Run("mainH", func(t *testing.T) {
		got := m.mainH()
		want := 48 // 50 - 2
		if got != want {
			t.Errorf("mainH() = %d, want %d", got, want)
		}
	})

	t.Run("listInnerW", func(t *testing.T) {
		got := m.listInnerW()
		want := 36 // 40 - 4
		if got != want {
			t.Errorf("listInnerW() = %d, want %d", got, want)
		}
	})

	t.Run("detailInnerW", func(t *testing.T) {
		got := m.detailInnerW()
		want := 56 // 60 - 4
		if got != want {
			t.Errorf("detailInnerW() = %d, want %d", got, want)
		}
	})

	t.Run("detailInnerH", func(t *testing.T) {
		got := m.detailInnerH()
		want := 46 // 48 - 2
		if got != want {
			t.Errorf("detailInnerH() = %d, want %d", got, want)
		}
	})
}

func TestLayoutHelpersDynamic(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		wantListW  int
	}{
		{"narrow terminal", 50, 19},        // 50*38/100=19, terminal too small for min clamp
		{"medium terminal", 120, 45},        // 120*38/100=45, within bounds
		{"wide terminal", 250, 80},          // 250*38/100=95, clamped to listMaxW=80
		{"very narrow", 30, 10},             // 30-20=10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{width: tt.width, height: 40}
			got := m.listW()
			if got != tt.wantListW {
				t.Errorf("listW() with width=%d: got %d, want %d", tt.width, got, tt.wantListW)
			}
			// detail should always be width - listW
			if m.detailW() != tt.width-got {
				t.Errorf("detailW() with width=%d: got %d, want %d", tt.width, m.detailW(), tt.width-got)
			}
		})
	}
}

func TestLayoutHelpersZeroSize(t *testing.T) {
	m := model{width: 0, height: 0}

	if m.listW() != 0 {
		t.Errorf("listW() with zero width = %d, want 0", m.listW())
	}
	if m.mainH() != -2 {
		t.Errorf("mainH() with zero height = %d, want -2", m.mainH())
	}
}
