package nex

import (
	"fmt"

	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	protocolglobals "github.com/PretendoNetwork/nex-protocols-go/v2/globals"
	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
)

type injAutoMatchHandler func(
	err error,
	packet nex.PacketInterface,
	callID uint32,
	criteria types.List[matchmakingtypes.MatchmakeSessionSearchCriteria],
	gathering matchmakingtypes.GatheringHolder,
	message types.String,
) (*nex.RMCMessage, *nex.Error)

// injAutoMatchProtocol patches only method 0x0F. INJ serializes this
// request with the pre-3.4 search-criteria/session layout (assumed from the
// NEX 3.2.1 library tags; verify on the first console session).
type injAutoMatchProtocol struct {
	endpoint nex.EndpointInterface
	handler  injAutoMatchHandler
}

func newINJAutoMatchProtocol(handler injAutoMatchHandler) *injAutoMatchProtocol {
	return &injAutoMatchProtocol{handler: handler}
}

func (protocol *injAutoMatchProtocol) Endpoint() nex.EndpointInterface {
	return protocol.endpoint
}

func (protocol *injAutoMatchProtocol) SetEndpoint(endpoint nex.EndpointInterface) {
	protocol.endpoint = endpoint
}

func injRequestLibraryVersions(endpoint nex.EndpointInterface) *nex.LibraryVersions {
	base := endpoint.LibraryVersions()
	versions := *base
	versions.MatchMaking = nex.NewLibraryVersion(3, 3, 0)
	return &versions
}

func decodeINJAutoMatchRequest(parameters []byte, endpoint nex.EndpointInterface) (
	types.List[matchmakingtypes.MatchmakeSessionSearchCriteria],
	matchmakingtypes.GatheringHolder,
	types.String,
	error,
) {
	criteria := types.NewList[matchmakingtypes.MatchmakeSessionSearchCriteria]()
	gathering := matchmakingtypes.NewGatheringHolder()
	message := types.NewString("")
	stream := nex.NewByteStreamIn(parameters, injRequestLibraryVersions(endpoint), endpoint.ByteStreamSettings())

	if err := criteria.ExtractFrom(stream); err != nil {
		return criteria, gathering, message, fmt.Errorf("read search criteria: %w", err)
	}
	if err := gathering.ExtractFrom(stream); err != nil {
		return criteria, gathering, message, fmt.Errorf("read gathering: %w", err)
	}
	if err := message.ExtractFrom(stream); err != nil {
		return criteria, gathering, message, fmt.Errorf("read message: %w", err)
	}
	if stream.Remaining() != 0 {
		return criteria, gathering, message, fmt.Errorf("unexpected trailing request data: %d bytes", stream.Remaining())
	}

	// This field does not exist in the pre-3.4 request layout. The generic common
	// handler expects it and the behavior of the older implementation was one
	// vacant participant when omitted.
	for index := range criteria {
		criteria[index].VacantParticipants = types.NewUInt16(1)
	}

	return criteria, gathering, message, nil
}

func (protocol *injAutoMatchProtocol) HandlePacket(packet nex.PacketInterface) {
	request := packet.RMCMessage()
	if request == nil || !request.IsRequest || request.ProtocolID != matchmakeextension.ProtocolID || request.MethodID != matchmakeextension.MethodAutoMatchmakeWithSearchCriteriaPostpone {
		return
	}

	criteria, gathering, message, err := decodeINJAutoMatchRequest(request.Parameters, packet.Sender().Endpoint())
	if err != nil {
		protocolglobals.Logger.Errorf("[INJ] AutoMatch request: %v (raw %d bytes: %x)", err, len(request.Parameters), request.Parameters)
		protocolglobals.RespondError(packet, matchmakeextension.ProtocolID, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error()))
		return
	}

	response, rmcError := protocol.handler(nil, packet, request.CallID, criteria, gathering, message)
	if rmcError != nil {
		protocolglobals.RespondError(packet, matchmakeextension.ProtocolID, rmcError)
		return
	}
	protocolglobals.Respond(packet, response)
}
