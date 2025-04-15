package source_test

import (
	"testing"
	"time"

	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Test(t *testing.T) { //nolint:funlen
	t.Parallel()

	now20250101, err := time.Parse("20060102150405", "20250101000000")
	require.NoError(t, err)

	newFilter := func(expression string, now time.Time) source.IDFilter {
		filter, err := source.NewIDFilter(expression, now)
		require.NoError(t, err)

		return filter
	}

	callTest := func(filter *source.IDFilter, id uint64) bool {
		result, err := filter.Test(id)
		require.NoError(t, err)

		return result
	}

	t.Run("simple boolean expressions", func(t *testing.T) {
		t.Parallel()

		t.Run("returns all IDs if expression is true", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("true", now20250101)

			assert.True(t, callTest(&filter, 20240322104533))
			assert.True(t, callTest(&filter, 20240413193712))
			assert.True(t, callTest(&filter, 20240629170508))
			assert.True(t, callTest(&filter, 20250715084320))
		})

		t.Run("returns no IDs if expression is false", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("false", now20250101)

			assert.False(t, callTest(&filter, 20240322104533))
			assert.False(t, callTest(&filter, 20240413193712))
			assert.False(t, callTest(&filter, 20240629170508))
			assert.False(t, callTest(&filter, 20250715084320))
		})
	})

	t.Run("filter by numerical ids", func(t *testing.T) {
		t.Parallel()

		t.Run("returns IDs before the given date", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("id < 20250101000000", now20250101)

			assert.True(t, callTest(&filter, 20241231235959))
			assert.True(t, callTest(&filter, 20240413193712))
			assert.True(t, callTest(&filter, 20240629170508))

			assert.False(t, callTest(&filter, 20250101000000))
			assert.False(t, callTest(&filter, 20250715084320))
		})

		t.Run("returns IDs in the given range", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("id >= 20250715084320 && id < 20251201132149", now20250101)

			assert.True(t, callTest(&filter, 20250715084320))
			assert.True(t, callTest(&filter, 20250910235402))
			assert.True(t, callTest(&filter, 20251130112944))

			assert.False(t, callTest(&filter, 20251201132149))
		})

		t.Run("returns IDs in the given list", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("id in [20240322104533, 20250910235402]", now20250101)

			assert.True(t, callTest(&filter, 20240322104533))
			assert.True(t, callTest(&filter, 20250910235402))

			assert.False(t, callTest(&filter, 20240413193712))
			assert.False(t, callTest(&filter, 20250715084320))
		})
	})

	t.Run("filter by string ids", func(t *testing.T) {
		t.Parallel()

		filter := newFilter(`hasPrefix(sid, "20240")`, now20250101)

		assert.True(t, callTest(&filter, 20240322104533))
		assert.True(t, callTest(&filter, 20240413193712))
		assert.True(t, callTest(&filter, 20240629170508))

		assert.False(t, callTest(&filter, 20250715084320))
	})

	t.Run("filter by date", func(t *testing.T) {
		t.Parallel()

		t.Run("returns IDs before the given date", func(t *testing.T) {
			t.Parallel()

			filter := newFilter(`createdAt < date("2025-01-01")`, now20250101)

			assert.True(t, callTest(&filter, 20241231235959))
			assert.True(t, callTest(&filter, 20240413193712))
			assert.True(t, callTest(&filter, 20240629170508))

			assert.False(t, callTest(&filter, 20250101000000))
			assert.False(t, callTest(&filter, 20250715084320))
		})

		t.Run("when a year is specified", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("createdAt.Year() == 2024", now20250101)

			assert.True(t, callTest(&filter, 20240322104533))
			assert.True(t, callTest(&filter, 20240413193712))

			assert.False(t, callTest(&filter, 20250715084320))
		})
	})

	t.Run("filter by age", func(t *testing.T) {
		t.Parallel()

		t.Run("returns IDs older than the given duration", func(t *testing.T) {
			t.Parallel()

			filter := newFilter(`age != nil && age < duration("24h")*280`, now20250101)

			assert.True(t, callTest(&filter, 20240413193712))
			assert.True(t, callTest(&filter, 20240629170508))

			assert.False(t, callTest(&filter, 20250715084320))
		})

		t.Run("returns IDs older than the given duration with ageDays alias", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("ageDays != nil && ageDays < 280", now20250101)

			assert.True(t, callTest(&filter, 20240413193712))
			assert.True(t, callTest(&filter, 20240629170508))

			assert.False(t, callTest(&filter, 20250715084320))
		})

		t.Run("nil age means the future", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("age == nil", now20250101)

			assert.True(t, callTest(&filter, 20250715084320))
			assert.True(t, callTest(&filter, 20250910235402))

			assert.False(t, callTest(&filter, 20240322104533))
		})
	})

	t.Run("uses the provdided now() function", func(t *testing.T) {
		t.Parallel()

		filter := newFilter("createdAt <= now()", now20250101)

		assert.True(t, callTest(&filter, 20250101000000))
		assert.True(t, callTest(&filter, 20240413193712))
		assert.True(t, callTest(&filter, 20240629170508))

		assert.False(t, callTest(&filter, 20250101000001))
		assert.False(t, callTest(&filter, 20251201132149))
	})

	t.Run("Edge cases", func(t *testing.T) {
		t.Parallel()

		t.Run("returns no IDs if expression is invalid", func(t *testing.T) {
			t.Parallel()

			_, err := source.NewIDFilter("invalid", now20250101)

			var compileErr *source.CompileFilterError

			require.ErrorAs(t, err, &compileErr)
			assert.Equal(t, "invalid", compileErr.Expression)
		})

		t.Run("fails when accessing unknown variable", func(t *testing.T) {
			t.Parallel()

			_, err := source.NewIDFilter("foobar == 1", now20250101)

			var compileErr *source.CompileFilterError

			require.ErrorAs(t, err, &compileErr)
			assert.Equal(t, "foobar == 1", compileErr.Expression)
		})

		t.Run("fails when expression returns non-bool", func(t *testing.T) {
			t.Parallel()

			_, err := source.NewIDFilter("52", now20250101)

			var compileErr *source.CompileFilterError

			require.ErrorAs(t, err, &compileErr)
			assert.Equal(t, "52", compileErr.Expression)
		})

		t.Run("fails when expression is incorrect", func(t *testing.T) {
			t.Parallel()

			_, err := source.NewIDFilter("id!!", now20250101)

			var compileErr *source.CompileFilterError

			require.ErrorAs(t, err, &compileErr)
			assert.Equal(t, "id!!", compileErr.Expression)
		})

		t.Run("NPE when run test", func(t *testing.T) {
			t.Parallel()

			filter := newFilter("ageDays < 280", now20250101)

			_, err := filter.Test(20250101000001)

			var runError *source.RunFilterError

			require.ErrorAs(t, err, &runError)
			assert.Equal(t, "ageDays < 280", runError.Expression)
			assert.Equal(t, uint64(20250101000001), runError.ID)
		})
	})
}
