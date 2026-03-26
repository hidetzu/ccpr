package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ParseResult
		wantErr bool
	}{
		{
			name:  "valid URL",
			input: "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/123",
			want: ParseResult{
				Region:     "ap-northeast-1",
				Repository: "my-repo",
				PRId:       "123",
			},
		},
		{
			name:  "valid URL with trailing slash",
			input: "https://us-east-1.console.aws.amazon.com/codesuite/codecommit/repositories/backend-api/pull-requests/42/",
			want: ParseResult{
				Region:     "us-east-1",
				Repository: "backend-api",
				PRId:       "42",
			},
		},
		{
			name:  "valid URL with query params",
			input: "https://eu-west-1.console.aws.amazon.com/codesuite/codecommit/repositories/web-app/pull-requests/7?region=eu-west-1",
			want: ParseResult{
				Region:     "eu-west-1",
				Repository: "web-app",
				PRId:       "7",
			},
		},
		{
			name:  "valid URL with fragment",
			input: "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/99#details",
			want: ParseResult{
				Region:     "ap-northeast-1",
				Repository: "my-repo",
				PRId:       "99",
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "not a URL",
			input:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "wrong host",
			input:   "https://github.com/user/repo/pull/1",
			wantErr: true,
		},
		{
			name:    "missing pull-requests segment",
			input:   "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo",
			wantErr: true,
		},
		{
			name:    "missing pr id",
			input:   "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/",
			wantErr: true,
		},
		{
			name:    "non-numeric pr id",
			input:   "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/abc",
			wantErr: true,
		},
		{
			name:    "missing repository name",
			input:   "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories//pull-requests/1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got %+v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
