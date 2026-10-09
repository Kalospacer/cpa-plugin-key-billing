package plugin

import (
	"encoding/json"
	"testing"

	"cpa-key-billing/internal/billing"
)

func TestHostAccountingDoesNotDependOnModelOrProviderNames(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want billing.TokenBreakdown
	}{
		{
			name: "subset cache included in input",
			raw:  `{"Provider":"mirasim","ExecutorType":"executorAdapter","Model":"same-model","Detail":{"InputTokens":100,"OutputTokens":50,"TotalTokens":150,"Breakdown":{"SchemaVersion":2,"Quality":"complete","TotalTokens":150,"Input":{"TotalTokens":100,"UncachedTokens":60,"CacheReadTokens":30,"CacheWriteTokens":10},"Output":{"TotalTokens":50,"NonReasoningTokens":30,"ReasoningTokens":20},"UnclassifiedTokens":0}}}`,
			want: billing.TokenBreakdown{Quality: billing.TokenAccountingComplete, TotalTokens: 150,
				Input:  billing.TokenInputBreakdown{TotalTokens: 100, UncachedTokens: 60, CacheReadTokens: 30, CacheWriteTokens: 10},
				Output: billing.TokenOutputBreakdown{TotalTokens: 50, NonReasoningTokens: 30, ReasoningTokens: 20}},
		},
		{
			name: "independent cache added to input",
			raw:  `{"Provider":"mirasim","ExecutorType":"executorAdapter","Model":"same-model","Detail":{"InputTokens":100,"OutputTokens":50,"TotalTokens":190,"Breakdown":{"SchemaVersion":2,"Quality":"complete","TotalTokens":190,"Input":{"TotalTokens":140,"UncachedTokens":100,"CacheReadTokens":30,"CacheWriteTokens":10},"Output":{"TotalTokens":50,"NonReasoningTokens":30,"ReasoningTokens":20},"UnclassifiedTokens":0}}}`,
			want: billing.TokenBreakdown{Quality: billing.TokenAccountingComplete, TotalTokens: 190,
				Input:  billing.TokenInputBreakdown{TotalTokens: 140, UncachedTokens: 100, CacheReadTokens: 30, CacheWriteTokens: 10},
				Output: billing.TokenOutputBreakdown{TotalTokens: 50, NonReasoningTokens: 30, ReasoningTokens: 20}},
		},
		{
			name: "separate reasoning added to output",
			raw:  `{"Provider":"mirasim","ExecutorType":"executorAdapter","Model":"same-model","Detail":{"InputTokens":100,"OutputTokens":50,"ReasoningTokens":20,"TotalTokens":170,"Breakdown":{"SchemaVersion":2,"Quality":"complete","TotalTokens":170,"Input":{"TotalTokens":100,"UncachedTokens":60,"CacheReadTokens":30,"CacheWriteTokens":10},"Output":{"TotalTokens":70,"NonReasoningTokens":50,"ReasoningTokens":20},"UnclassifiedTokens":0}}}`,
			want: billing.TokenBreakdown{Quality: billing.TokenAccountingComplete, TotalTokens: 170,
				Input:  billing.TokenInputBreakdown{TotalTokens: 100, UncachedTokens: 60, CacheReadTokens: 30, CacheWriteTokens: 10},
				Output: billing.TokenOutputBreakdown{TotalTokens: 70, NonReasoningTokens: 50, ReasoningTokens: 20}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record UsageRecord
			if err := json.Unmarshal([]byte(tc.raw), &record); err != nil {
				t.Fatal(err)
			}
			if got := usageBreakdown(record); got != tc.want {
				t.Fatalf("host accounting was reinterpreted: got %+v want %+v", got, tc.want)
			}
		})
	}
}

func TestHostAccountingRejectsUnknownSchemaAndNegativeBuckets(t *testing.T) {
	for _, invalid := range []*UsageTokenBreakdown{
		{SchemaVersion: 3, Quality: "complete"},
		{SchemaVersion: 2, Quality: "unknown", TotalTokens: 5},
		{SchemaVersion: 2, Quality: "complete", Input: UsageTokenInputBreakdown{UncachedTokens: -1}},
	} {
		if result, ok := hostTokenBreakdown(invalid); ok {
			t.Fatalf("accepted invalid host contract: %+v -> %+v", invalid, result)
		}
	}
}
