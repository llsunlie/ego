package ai

import "testing"

func TestParseConstellationBorderlineJudgeJSON_PrimaryAndSecondary(t *testing.T) {
	resp, err := parseConstellationBorderlineJudgeJSON(`{
  "decision": "use_existing",
  "primary": {
    "constellation_id": "c-main",
    "theme_code": "theme_main",
    "confidence": 0.82,
    "shared_theme": "入职流程反复卡住",
    "match_dimensions": ["situation"],
    "reason": "这是主归属"
  },
  "secondary": [
    {
      "constellation_id": "c-side",
      "theme_code": "theme_side",
      "confidence": 0.72,
      "shared_theme": "也有被审核的位置感",
      "match_dimensions": ["identity"],
      "reason": "这是副视角"
    }
  ],
  "suggested_theme_code": "",
  "suggested_theme_label": "",
  "suggested_theme_description": ""
}`)
	if err != nil {
		t.Fatalf("parseConstellationBorderlineJudgeJSON() error = %v", err)
	}
	if resp.Decision != "use_existing" {
		t.Fatalf("decision = %q", resp.Decision)
	}
	if resp.Primary == nil || resp.Primary.ConstellationID != "c-main" {
		t.Fatalf("primary = %#v", resp.Primary)
	}
	if len(resp.Secondary) != 1 || resp.Secondary[0].ConstellationID != "c-side" {
		t.Fatalf("secondary = %#v", resp.Secondary)
	}
}

func TestParseConstellationBorderlineJudgeJSON_LegacyShape(t *testing.T) {
	resp, err := parseConstellationBorderlineJudgeJSON(`{
  "decision": "use_existing",
  "constellation_id": "c-main",
  "theme_code": "theme_main",
  "confidence": 0.82,
  "shared_situation": "入职流程反复卡住",
  "match_dimensions": ["situation"],
  "reason": "这是主归属"
}`)
	if err != nil {
		t.Fatalf("parseConstellationBorderlineJudgeJSON() error = %v", err)
	}
	if resp.Decision != "use_existing" || resp.ConstellationID != "c-main" {
		t.Fatalf("legacy response = %#v", resp)
	}
}
