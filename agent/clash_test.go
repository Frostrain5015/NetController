package main

import (
	"math"
	"testing"
)

func TestSubscriptionTextUsageInfersNearlyFullPlanCap(t *testing.T) {
	usage := subscriptionUsage{}
	usage.mergeText("剩余流量：149.98 GB")
	usage.mergeText("套餐到期：2026-12-16")
	usage.inferTotalFromRemaining()

	assertFloatPtr(t, usage.RemainingGB, 149.98)
	assertFloatPtr(t, usage.TotalGB, 150)
	assertFloatPtr(t, usage.UsedGB, 0.02)
	if usage.Expiry != "2026-12-16" {
		t.Fatalf("expiry = %q, want %q", usage.Expiry, "2026-12-16")
	}
}

func TestSubscriptionTextUsageDoesNotInferDistantCap(t *testing.T) {
	usage := subscriptionUsage{}
	usage.mergeText("剩余流量：174.34 GB")
	usage.inferTotalFromRemaining()

	assertFloatPtr(t, usage.RemainingGB, 174.34)
	if usage.TotalGB != nil {
		t.Fatalf("total = %.2f, want nil", *usage.TotalGB)
	}
	if usage.UsedGB != nil {
		t.Fatalf("used = %.2f, want nil", *usage.UsedGB)
	}
}

func TestSubscriptionTextUsageParsesExplicitTotal(t *testing.T) {
	usage := subscriptionUsage{}
	usage.mergeText("总流量：150 GB")

	assertFloatPtr(t, usage.TotalGB, 150)
}

func TestAutoSelectProxyGroupCanBeListed(t *testing.T) {
	if !isAutoSelectProxy("♻️ 自动选择", "urltest") {
		t.Fatal("auto select urltest group should be listable")
	}
	if isAutoSelectProxy("🎬 NETFLIX", "selector") {
		t.Fatal("ordinary selector group should not be treated as auto select")
	}
}

func assertFloatPtr(t *testing.T, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("got nil, want %.2f", want)
	}
	if math.Abs(*got-want) > 0.01 {
		t.Fatalf("got %.4f, want %.4f", *got, want)
	}
}
