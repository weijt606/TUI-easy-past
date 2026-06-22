package detect

import "testing"

func TestDetect(t *testing.T) {
	cases := []struct {
		name string
		text string
		want Format
	}{
		{
			name: "plain prose",
			text: "This is just a normal paragraph of text.\nNothing special about it at all here.",
			want: PlainText,
		},
		{
			name: "heading",
			text: "# Release notes\n\nWe shipped a thing today.",
			want: Markdown,
		},
		{
			name: "fenced code",
			text: "Run this:\n\n```sh\nnpm install\n```\n",
			want: Markdown,
		},
		{
			name: "bullet list",
			text: "Todo:\n- buy milk\n- walk dog\n- write code",
			want: Markdown,
		},
		{
			name: "table",
			text: "| a | b |\n|---|---|\n| 1 | 2 |",
			want: Markdown,
		},
		{
			name: "prose with a single dash sentence",
			text: "I went to the store - it was closed.\nSo I came back home again.",
			want: PlainText,
		},
		{
			name: "links and bold",
			text: "See the [docs](https://x.y) for **important** details.",
			want: Markdown,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Detect(tc.text)
			if got.Format != tc.want {
				t.Errorf("Detect() = %v (score=%.1f signals=%v), want %v",
					got.Format, got.Score, got.Signals, tc.want)
			}
		})
	}
}
