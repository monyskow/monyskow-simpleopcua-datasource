package plugin

import (
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/opcua"
)

// ---------------------------------------------------------------------------
// Fix #4 — buildDataFrame timestamp priority
// ---------------------------------------------------------------------------

// newBareDataSource returns the minimal Datasource needed to call buildDataFrame.
// buildDataFrame only reads fields, so no client or certMgr is needed.
func newBareDataSource() *Datasource {
	return &Datasource{
		logger: log.DefaultLogger,
	}
}

func frameTimestamp(t *testing.T, d *Datasource, nodes []models.NodeQuery, values []opcua.NodeValue) time.Time {
	t.Helper()
	frame := d.buildDataFrame(nodes, values)
	require.NotNil(t, frame)
	require.NotEmpty(t, frame.Fields, "frame must have at least the time field")
	ts, ok := frame.Fields[0].At(0).(time.Time)
	require.True(t, ok, "first field first value must be time.Time")
	return ts
}

func TestBuildDataFrame_SourceTimestamp_UsesSourceTimestamp(t *testing.T) {
	// Fix #4: when SourceTimestamp is non-zero it must be preferred.
	d := newBareDataSource()
	want := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	values := []opcua.NodeValue{{
		NodeID:          "ns=1;i=1",
		SourceTimestamp: want,
		ServerTimestamp: want.Add(time.Hour), // different — must NOT be used
	}}
	nodes := []models.NodeQuery{{NodeID: "ns=1;i=1", DisplayName: "Temp"}}

	got := frameTimestamp(t, d, nodes, values)

	assert.Equal(t, want, got, "SourceTimestamp must be preferred over ServerTimestamp")
}

func TestBuildDataFrame_ZeroSourceTimestamp_UsesServerTimestamp(t *testing.T) {
	// Fix #4: when SourceTimestamp is zero but ServerTimestamp is set, use ServerTimestamp.
	d := newBareDataSource()
	want := time.Date(2024, 6, 1, 15, 30, 0, 0, time.UTC)

	values := []opcua.NodeValue{{
		NodeID:          "ns=1;i=2",
		SourceTimestamp: time.Time{}, // zero
		ServerTimestamp: want,
	}}
	nodes := []models.NodeQuery{{NodeID: "ns=1;i=2", DisplayName: "Pressure"}}

	got := frameTimestamp(t, d, nodes, values)

	assert.Equal(t, want, got, "ServerTimestamp must be used when SourceTimestamp is zero")
}

func TestBuildDataFrame_BothTimestampsZero_UsesNow(t *testing.T) {
	// Fix #4: when both timestamps are zero the frame should use wall-clock time.
	d := newBareDataSource()
	before := time.Now()

	values := []opcua.NodeValue{{
		NodeID:          "ns=1;i=3",
		SourceTimestamp: time.Time{},
		ServerTimestamp: time.Time{},
	}}
	nodes := []models.NodeQuery{{NodeID: "ns=1;i=3", DisplayName: "Flow"}}

	got := frameTimestamp(t, d, nodes, values)
	after := time.Now()

	assert.False(t, got.Before(before), "timestamp must not be before the call")
	assert.False(t, got.After(after.Add(time.Second)), "timestamp must be within 1s of now")
}

func TestBuildDataFrame_EmptyValues_UsesNow(t *testing.T) {
	// Fix #4 edge case: empty values slice must not panic and must fall through to Now().
	d := newBareDataSource()
	nodes := []models.NodeQuery{{NodeID: "ns=1;i=4", DisplayName: "Speed"}}
	before := time.Now()

	var got time.Time
	require.NotPanics(t, func() {
		got = frameTimestamp(t, d, nodes, []opcua.NodeValue{})
	})

	assert.False(t, got.Before(before), "timestamp must not be before the call")
	assert.WithinDuration(t, time.Now(), got, time.Second)
}

func TestBuildDataFrame_NilValues_UsesNow(t *testing.T) {
	// Nil slice is equivalent to empty: must not panic.
	d := newBareDataSource()
	nodes := []models.NodeQuery{{NodeID: "ns=1;i=5", DisplayName: "Vibration"}}

	require.NotPanics(t, func() {
		frame := d.buildDataFrame(nodes, nil)
		require.NotNil(t, frame)
		require.NotEmpty(t, frame.Fields)
	})
}

// ---------------------------------------------------------------------------
// Fix #2 — Dispose() / getOrCreateClient Close() error logging
//
// There is no injectable seam for the inner OPC UA client, so we cannot force
// Close() to return an error in a unit test without refactoring production code
// (which is developer work).  What we CAN verify: Dispose() on a Datasource
// with a nil client does not panic and leaves the struct in a clean state.
// ---------------------------------------------------------------------------

func TestDispose_NilClient_DoesNotPanic(t *testing.T) {
	d := &Datasource{
		logger:   log.DefaultLogger,
		settings: models.DataSourceSettings{Timeout: 5},
	}
	// d.client is nil — this exercises the nil-guard branch, not the warn-on-close path.
	require.NotPanics(t, d.Dispose)
}

func TestDispose_Idempotent_NilClient(t *testing.T) {
	d := &Datasource{
		logger:   log.DefaultLogger,
		settings: models.DataSourceSettings{Timeout: 5},
	}
	require.NotPanics(t, d.Dispose)
	require.NotPanics(t, d.Dispose, "second Dispose must not panic")
}
