package database

import (
	"database/sql"
	"errors"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
)

// INJGetPersistenceInfo answers DataStore::GetPersistenceInfo (the retail
// client calls it during a match): which object is pinned to (owner, slot).
// Objects are pinned by INJPerpetuateObject.
func INJGetPersistenceInfo(err error, packet nex.PacketInterface, callID uint32, ownerID types.PID, persistenceSlotID types.UInt16) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, err.Error())
	}

	var dataID int64
	scanErr := Postgres.QueryRow(`
		SELECT data_id FROM datastore.objects
		WHERE owner = $1 AND persistence_slot_id = $2 AND deleted = false
		ORDER BY update_date DESC LIMIT 1`,
		int64(ownerID), int64(persistenceSlotID)).Scan(&dataID)
	if errors.Is(scanErr, sql.ErrNoRows) {
		return nil, nex.NewError(nex.ResultCodes.DataStore.NotFound, "not found")
	}
	if scanErr != nil {
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, scanErr.Error())
	}

	info := datastore_types.NewDataStorePersistenceInfo()
	info.OwnerID = ownerID
	info.PersistenceSlotID = persistenceSlotID
	info.DataID = types.NewUInt64(uint64(dataID))

	endpoint := packet.Sender().Endpoint()
	stream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	info.WriteTo(stream)

	rmcResponse := nex.NewRMCSuccess(endpoint, stream.Bytes())
	rmcResponse.ProtocolID = datastore.ProtocolID
	rmcResponse.MethodID = datastore.MethodGetPersistenceInfo
	rmcResponse.CallID = callID

	return rmcResponse, nil
}
