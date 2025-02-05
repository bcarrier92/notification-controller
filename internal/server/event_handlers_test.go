package server

import (
	"errors"
	"strings"
	"testing"

	"github.com/fluxcd/pkg/runtime/logger"
)

func TestRedactTokenFromError(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		originalErrStr string
		expectedErrStr string
	}{
		{
			name:           "no token",
			token:          "8h0387hdyehbwwa45",
			originalErrStr: "Cannot post to github",
			expectedErrStr: "Cannot post to github",
		},
		{
			name:           "empty token",
			token:          "",
			originalErrStr: "Cannot post to github",
			expectedErrStr: "Cannot post to github",
		},
		{
			name:           "exact token",
			token:          "8h0387hdyehbwwa45",
			originalErrStr: "Cannot post to github with token 8h0387hdyehbwwa45",
			expectedErrStr: "Cannot post to github with token *****",
		},
		{
			name:           "non-exact token",
			token:          "8h0387hdyehbwwa45",
			originalErrStr: `Cannot post to github with token 8h0387hdyehbwwa45\\n`,
			expectedErrStr: `Cannot post to github with token *****\\n`,
		},
		{
			name:           "extra text in front token",
			token:          "8h0387hdyehbwwa45",
			originalErrStr: `Cannot post to github with token metoo8h0387hdyehbwwa45\\n`,
			expectedErrStr: `Cannot post to github with token metoo*****\\n`,
		},
		{
			name:           "extra text in front token",
			token:          "8h0387hdyehbwwa45踙",
			originalErrStr: `Cannot post to github with token metoo8h0387hdyehbwwa45踙\\n`,
			expectedErrStr: `Cannot post to github with token metoo*****\\n`,
		},
		{
			name:           "return error on invalid UTF-8 string",
			token:          "\x18\xd0\xfa\xab\xb2\x93\xbb;\xc0l\xf4\xdc",
			originalErrStr: `Cannot post to github with token \x18\xd0\xfa\xab\xb2\x93\xbb;\xc0l\xf4\xdc\\n`,
			expectedErrStr: `error redacting token from error message`,
		},
		{
			name:           "unescaped token",
			token:          "8h0387hdyehbwwa45\\",
			originalErrStr: `Cannot post to github with token metoo8h0387hdyehbwwa45\\\n`,
			expectedErrStr: `Cannot post to github with token metoo*****n`,
		},
		{
			name:           "invalid chars",
			token:          "8h0387hdyehbwwa45(?!\\/)",
			originalErrStr: `Cannot post to github`,
			expectedErrStr: `Cannot post to github`,
		},
	}

	for _, tt := range tests {
		log := logger.NewLogger(logger.Options{})
		err := redactTokenFromError(errors.New(tt.originalErrStr), tt.token, log)
		if err == nil {
			t.Fatalf("error shouldn't be nil")
		}

		if !strings.Contains(err.Error(), tt.expectedErrStr) {
			t.Errorf("expected error string '%s' but got '%s'",
				tt.expectedErrStr, err)
		}
	}

}
