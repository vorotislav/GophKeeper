package logger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		giveLevel   string
		giveFormat  string
		wantErrFunc require.ErrorAssertionFunc
		wantErr     bool
	}{
		{
			name:      "invalid level",
			giveLevel: "invalid",
			wantErr:   true,
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrUnsupportedLogLevel)
			},
		},
		{
			name:       "invalid format",
			giveLevel:  "debug",
			giveFormat: "invalid",
			wantErr:    true,
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrUnsupportedLogFormat)
			},
		},
		{
			name:       "success",
			giveLevel:  "debug",
			giveFormat: LogFormatJSON,
			wantErr:    false,
			wantErrFunc: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := New(tc.giveLevel, tc.giveFormat, "stdout", true)
			tc.wantErrFunc(t, err)
			if !tc.wantErr {
				require.NotNil(t, got)
			}
		})
	}
}
