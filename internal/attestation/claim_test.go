package attestation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAdvisoriesFromTCBInfo(t *testing.T) {
	const tcbInfo = `
{
	"tcbInfo": {
		"tcbLevels": [
			{
				"tcbStatus": "UpToDate"
			},
			{
				"tcbStatus": "OutOfDate",
				"advisoryIDs": [
					"id 1"
				]
			},
			{
				"tcbStatus": "OutOfDate",
				"advisoryIDs": [
					"id 2",
					"id 3",
					"id 4"
				]
			}
		]
	}
}
`

	testCases := map[string]struct {
		info           string
		index          uint
		wantErr        bool
		wantAdvisories []string
	}{
		"0 advisories at index 0": {
			info:  tcbInfo,
			index: 0,
		},
		"1 advisory at index 1": {
			info:           tcbInfo,
			index:          1,
			wantAdvisories: []string{"id 1"},
		},
		"3 advisories at index 2": {
			info:           tcbInfo,
			index:          2,
			wantAdvisories: []string{"id 2", "id 3", "id 4"},
		},
		"error at index 3": {
			info:    tcbInfo,
			index:   3,
			wantErr: true,
		},
		"no TCB info": {
			info:    "",
			index:   0,
			wantErr: true,
		},
		"null-terminated TCB info": {
			info:           tcbInfo + "\x00",
			index:          1,
			wantAdvisories: []string{"id 1"},
		},
		"multiple null terminators": {
			info:           tcbInfo + "\x00\x00\x00",
			index:          1,
			wantAdvisories: []string{"id 1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			advisories, err := getAdvisoriesFromTCBInfo([]byte(tc.info), tc.index)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.wantAdvisories, advisories)
		})
	}
}
