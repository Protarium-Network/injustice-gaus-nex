// Package nex is the Injustice: Gods Among Us (Wii U) NEX server (game_server_id
// 10144800 / 269995264). The retail RPX statically links NEX and drives it
// through nn::act::AcquireNexServiceToken with that game server ID - the
// standard Wii U NASC -> nex_token -> PRUDP path. Authentication and secure
// run as separate PRUDP endpoints.
package nex

import (
	"fmt"
	"os"
	"strconv"

	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_ticket_granting "github.com/PretendoNetwork/nex-protocols-common-go/v2/ticket-granting"
	ticket_granting "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting"
	"github.com/Protarium-Network/injustice-gaus-nex/globals"
)

var AuthenticationServer *nex.PRUDPServer
var AuthenticationEndpoint *nex.PRUDPEndPoint

func StartAuthenticationServer() {
	AuthenticationServer = nex.NewPRUDPServer()

	// Wrong value = 106-0502 / endless CONNECT retransmit. Most 3.4.13 titles
	// on this stack need false; flip with PN_INJ_LEGACY_SIGNATURE=1.
	AuthenticationServer.PRUDPV1Settings.LegacyConnectionSignature = globals.LegacyConnectionSignature

	AuthenticationEndpoint = nex.NewPRUDPEndPoint(1)
	AuthenticationEndpoint.ServerAccount = globals.AuthenticationServerAccount
	AuthenticationEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	AuthenticationEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername
	AuthenticationServer.BindPRUDPEndPoint(AuthenticationEndpoint)

	AuthenticationServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(globals.NEXMajor, globals.NEXMinor, globals.NEXPatch))
	AuthenticationServer.AccessKey = globals.AccessKey
	// NEX 3.4.x titles do not write the structure-header version byte. Flip to
	// true on both endpoints if LoginEx fails with a structure length error.
	AuthenticationServer.ByteStreamSettings.UseStructureHeader = false

	AuthenticationEndpoint.OnData(func(packet nex.PacketInterface) {
		request := packet.RMCMessage()
		if request == nil {
			return
		}
		pid := uint64(packet.Sender().PID())
		fmt.Printf("[INJ Auth] PID=%d protocol=0x%02X method=0x%02X\n", pid, request.ProtocolID, request.MethodID)
		// One-off wire-format diagnostic: dump the raw RMC parameter bytes
		// for LoginEx so the AuthenticationInfo layout can be confirmed by
		// hand against a real capture instead of guessing.
		if request.ProtocolID == 0x0A && request.MethodID == 0x02 {
			fmt.Printf("[INJ Auth] RAW LoginEx params (%d bytes): %x\n", len(request.Parameters), request.Parameters)
		}
	})

	registerAuthenticationServerProtocols()

	port, _ := strconv.Atoi(os.Getenv("PN_INJ_AUTH_PORT"))
	globals.Logger.Successf("[INJ] Authentication server listening on UDP %d", port)
	AuthenticationServer.Listen(port)
}

func registerAuthenticationServerProtocols() {
	ticketGrantingProtocol := ticket_granting.NewProtocol()
	AuthenticationEndpoint.RegisterServiceProtocol(ticketGrantingProtocol)
	commonTicketGrantingProtocol := common_ticket_granting.NewCommonProtocol(ticketGrantingProtocol)

	securePort, _ := strconv.Atoi(os.Getenv("PN_INJ_SECURE_PORT"))

	// Must stay short (~15 chars): the retail binary truncates this field
	// into a small fixed-size buffer, so use a bare host, not a subdomain.
	secureHost := os.Getenv("PN_INJ_SECURE_HOST")
	if secureHost == "" {
		secureHost = "localhost"
	}

	secureStationURL := types.NewStationURL("")
	secureStationURL.SetURLType(constants.StationURLPRUDPS)
	secureStationURL.SetAddress(secureHost)
	secureStationURL.SetPortNumber(uint16(securePort))
	secureStationURL.SetConnectionID(1)
	secureStationURL.SetPrincipalID(types.NewPID(2))
	secureStationURL.SetStreamID(1)
	secureStationURL.SetStreamType(constants.StreamTypeRVSecure)
	secureStationURL.SetType(uint8(constants.StationURLFlagPublic))

	commonTicketGrantingProtocol.ValidateLoginData = globals.ValidateLoginData
	commonTicketGrantingProtocol.SecureStationURL = secureStationURL
	commonTicketGrantingProtocol.BuildName = types.NewString("")
	commonTicketGrantingProtocol.SecureServerAccount = globals.SecureServerAccount
}
