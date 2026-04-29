package billing_plan

const (
	FEATURE_DISABLED int64 = 0
	FEATURE_ENABLED  int64 = 1
)

type FeatureSet struct {
	CustomBranding    int64 `json:"custom_branding,omitempty"`
	AdvancedAnalytics int64 `json:"advanced_analytics,omitempty" public:"view"`
	PrioritySupport   int64 `json:"priority_support,omitempty"`
}

type MergeableFeatureSet struct{}

func (this *FeatureSet) Merge(_ *MergeableFeatureSet) {
}
