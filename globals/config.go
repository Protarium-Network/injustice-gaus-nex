package globals

import (
	"os"
	"strings"
)

// NEX configuration for "Injustice: Gods Among Us" (Wii U, NetherRealm Studios /
// Warner Bros. Interactive, 2013).
//
// AccessKey and game server ID come from the retail executable
// (DCFGame-Retail.rpx, see RECON.md) and are cross-checked against
// kinnay.github.io's public Wii U NEX database (id 269555456, key 65e9f4d6,
// branch origin/release/ngs/3.4.x.3, build 3_4_13_3_0).
//
// The game ships under one title ID per region; each region's game server ID
// is the unique-ID half of its title ID. The PRUDP server is region-agnostic
// (one shared access key); the IDs only matter to the NASC / nex_token
// front-end, which must map EVERY ID below to this server's auth port.
const (
	AccessKey = "65e9f4d6"

	GameServerIDUSA = "10111700" // title 0005000010111700 (WUP-P-AJSE) - in the binary
	GameServerIDEUR = "10111a00" // title 0005000010111A00 (WUP-P-AJSP)
	GameServerIDJPN = "10140700" // title 0005000010140700 (WUP-P-AJSJ)
)

// GameServerIDs lists every region this server is meant to serve.
var GameServerIDs = []string{GameServerIDUSA, GameServerIDEUR, GameServerIDJPN}

// kinnay build tag "build:3_4_13_3_0" (origin/release/ngs/3.4.x.3), same NGS
// generation as Trine 2 / Sochi 2014 / DKC:TF.
//
// MatchMaking is pinned to 3.3.0 (pre-3.4 MatchmakeSession layout: no
// ProgressScore byte, as console-captured on Trine 2 and Sochi 2014). Ranking
// is bumped to 3.6.0 in nex/secure.go so RankingRankData.UpdateTime is
// serialized.
const (
	NEXMajor = 3
	NEXMinor = 4
	NEXPatch = 13

	MatchMakingMajor = 3
	MatchMakingMinor = 3
	MatchMakingPatch = 0
)

// LegacyConnectionSignature selects the PRUDPv1 CONNECT-ACK signature scheme.
// Unverified on a console for this title: most 3.4.13 Wii U titles here need
// false (Sonic Lost World, DKC:TF), Trine 2 needed true. Set
// PN_INJ_LEGACY_SIGNATURE=1 if the console reports 106-0502.
var LegacyConnectionSignature = false

func init() {
	if v := strings.TrimSpace(os.Getenv("PN_INJ_LEGACY_SIGNATURE")); v != "" {
		LegacyConnectionSignature = v == "1"
	}
}
