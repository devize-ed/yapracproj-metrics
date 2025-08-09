package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	type args struct {
		data []byte
		key  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "hash_data",
			args: args{
				data: []byte("test_data"),
				key:  "test_key",
			},
		},
		{
			name: "hash_empty_key",
			args: args{
				data: []byte("test_data"),
				key:  "",
			},
		},
	}
	for _, tt := range tests {
		h := hmac.New(sha256.New, []byte(tt.args.key))
		h.Write(tt.args.data)
		want := hex.EncodeToString(h.Sum(nil))

		got := Hash(tt.args.data, tt.args.key)
		assert.Equal(t, want, got)
		assert.Equal(t, 64, len(got))
	}
}

func TestVerify(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	type args struct {
		data []byte
		key  string
		hash string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "verify_correct",
			args: args{
				data: []byte("test_data"),
				key:  "test_key",
				hash: Hash([]byte("test_data"), "test_key"),
			},
			want: true,
		},
		{
			name: "verify_wrong",
			args: args{
				data: []byte("test_data"),
				key:  "test_key",
				hash: hex.EncodeToString([]byte("wrong_hash")),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := Verify(tt.args.data, tt.args.key, tt.args.hash)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, ok)
		})
	}
}
