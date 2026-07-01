package main

// Static mappings between Branch names (frontend) and UUIDs (backend DB)
var branchToUUID = map[string]string{
	"Head Office":      "00000000-0000-0000-0000-000000000001",
	"Jakarta Branch":   "00000000-0000-0000-0000-000000000002",
	"Surabaya Branch":  "00000000-0000-0000-0000-000000000003",
	"Bandung Branch":   "00000000-0000-0000-0000-000000000004",
	"Bali Branch":      "00000000-0000-0000-0000-000000000005",
}

var uuidToBranch = map[string]string{
	"00000000-0000-0000-0000-000000000001": "Head Office",
	"00000000-0000-0000-0000-000000000002": "Jakarta Branch",
	"00000000-0000-0000-0000-000000000003": "Surabaya Branch",
	"00000000-0000-0000-0000-000000000004": "Bandung Branch",
	"00000000-0000-0000-0000-000000000005": "Bali Branch",
}
