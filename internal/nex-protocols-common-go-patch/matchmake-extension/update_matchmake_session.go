package matchmake_extension

import (
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	match_making_types "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	matchmake_extension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	pqextended "github.com/PretendoNetwork/pq-extended"
)

// updateMatchmakeSession lets the host rewrite its session (game mode,
// attributes, open participation, application buffer, size limits). The
// retail Injustice client calls it as soon as a player joins; answering
// NotImplemented made it report "the game session is no longer available".
func (commonProtocol *CommonProtocol) updateMatchmakeSession(err error, packet nex.PacketInterface, callID uint32, anyGathering match_making_types.GatheringHolder) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	if !anyGathering.Object.GatheringObjectID().Equals(types.NewString("MatchmakeSession")) {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}
	session := anyGathering.Object.(match_making_types.MatchmakeSession)

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)
	pid := uint64(connection.PID())

	attribs := make([]uint32, len(session.Attributes))
	for i, value := range session.Attributes {
		attribs[i] = uint32(value)
	}

	commonProtocol.manager.Mutex.Lock()
	defer commonProtocol.manager.Mutex.Unlock()

	var ownerPID, hostPID uint64
	row := commonProtocol.manager.Database.QueryRow(`SELECT owner_pid, host_pid FROM matchmaking.gatherings WHERE id=$1 AND registered=true`, uint32(session.Gathering.ID))
	if scanErr := row.Scan(&ownerPID, &hostPID); scanErr != nil {
		return nil, nex.NewError(nex.ResultCodes.RendezVous.SessionVoid, "change_error")
	}
	if pid != ownerPID && pid != hostPID {
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "change_error")
	}

	if _, execErr := commonProtocol.manager.Database.Exec(`UPDATE matchmaking.matchmake_sessions SET
		game_mode=$2, attribs=$3, open_participation=$4, application_buffer=$5
		WHERE id=$1`,
		uint32(session.Gathering.ID), uint32(session.GameMode), pqextended.Array(attribs), bool(session.OpenParticipation), []byte(session.ApplicationBuffer),
	); execErr != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.Unknown, execErr.Error())
	}

	if _, execErr := commonProtocol.manager.Database.Exec(`UPDATE matchmaking.gatherings SET
		min_participants=$2, max_participants=$3, participation_policy=$4, policy_argument=$5, flags=$6, description=$7
		WHERE id=$1`,
		uint32(session.Gathering.ID), uint16(session.Gathering.MinimumParticipants), uint16(session.Gathering.MaximumParticipants),
		uint32(session.Gathering.ParticipationPolicy), uint32(session.Gathering.PolicyArgument), uint32(session.Gathering.Flags), string(session.Gathering.Description),
	); execErr != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.Unknown, execErr.Error())
	}

	fmt.Printf("[INJ MM] UpdateMatchmakeSession GID=%d by PID=%d attribs=%v open=%t max=%d\n",
		uint32(session.Gathering.ID), pid, attribs, bool(session.OpenParticipation), uint16(session.Gathering.MaximumParticipants))

	rmcResponse := nex.NewRMCSuccess(endpoint, nil)
	rmcResponse.ProtocolID = matchmake_extension.ProtocolID
	rmcResponse.MethodID = matchmake_extension.MethodUpdateMatchmakeSession
	rmcResponse.CallID = callID

	return rmcResponse, nil
}
