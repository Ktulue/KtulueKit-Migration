package mapper_test

import (
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/mapper"
)

func TestApplyDestOverride(t *testing.T) {
	cases := []struct {
		name   string
		target string
		dest   string
		want   string
	}{
		// Guard 1: empty destRoot → unchanged
		{
			name:   "empty dest returns target unchanged",
			target: `C:\Users\Foo\AppData\Roaming\App`,
			dest:   "",
			want:   `C:\Users\Foo\AppData\Roaming\App`,
		},
		// Guard 2: target has no drive prefix → unchanged
		{
			name:   "relative target returned unchanged",
			target: `relative\path`,
			dest:   `D:\`,
			want:   `relative\path`,
		},
		{
			name:   "target too short to have drive prefix → unchanged",
			target: `C:`,
			dest:   `D:\`,
			want:   `C:`,
		},
		// Guard 3: destRoot is drive root (len==3) → drive swap
		{
			name:   "drive swap on full path",
			target: `C:\Users\Foo\AppData\Roaming\App`,
			dest:   `D:\`,
			want:   `D:\Users\Foo\AppData\Roaming\App`,
		},
		{
			name:   "drive swap on root only",
			target: `C:\`,
			dest:   `D:\`,
			want:   `D:\`,
		},
		// Guard 4: destRoot is longer path → prefix substitution
		{
			name:   "prefix substitution",
			target: `C:\Users\Foo\AppData\Roaming\App`,
			dest:   `D:\Restored\`,
			want:   `D:\Restored\Users\Foo\AppData\Roaming\App`,
		},
		{
			name:   "prefix substitution strips only X:\\",
			target: `C:\SomeApp`,
			dest:   `E:\Backup\New\`,
			want:   `E:\Backup\New\SomeApp`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapper.ApplyDestOverride(tc.target, tc.dest)
			if got != tc.want {
				t.Errorf("ApplyDestOverride(%q, %q) = %q; want %q", tc.target, tc.dest, got, tc.want)
			}
		})
	}
}
