package database

import (
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
)

// INJPerpetuateObject answers DataStore::PerpetuateObject. The retail client
// creates its profile objects with PostMetaBinary and then pins each one to a
// persistence slot with this call; it later reads them back by
// (owner, slot) through GetMeta, so the object must end up in that slot.
// deleteLastObject drops whatever the slot held before.
func INJPerpetuateObject(err error, packet nex.PacketInterface, callID uint32, persistenceSlotID types.UInt16, dataID types.UInt64, deleteLastObject types.Bool) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, err.Error())
	}

	owner := int64(packet.Sender().PID())

	if bool(deleteLastObject) {
		if _, dbErr := Postgres.Exec(`
			UPDATE datastore.objects SET deleted = true
			WHERE owner = $1 AND persistence_slot_id = $2 AND data_id <> $3 AND deleted = false`,
			owner, int64(persistenceSlotID), int64(dataID)); dbErr != nil {
			return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, "perpetuate: "+dbErr.Error())
		}
	}

	result, dbErr := Postgres.Exec(`
		UPDATE datastore.objects SET persistence_slot_id = $1
		WHERE data_id = $2 AND owner = $3 AND deleted = false`,
		int64(persistenceSlotID), int64(dataID), owner)
	if dbErr != nil {
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, "perpetuate: "+dbErr.Error())
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return nil, nex.NewError(nex.ResultCodes.DataStore.NotFound, "not found")
	}

	fmt.Printf("[INJ DataStore] PerpetuateObject slot=%d dataID=%d deleteLast=%t\n", uint16(persistenceSlotID), uint64(dataID), bool(deleteLastObject))

	endpoint := packet.Sender().Endpoint()
	rmcResponse := nex.NewRMCSuccess(endpoint, nil)
	rmcResponse.ProtocolID = datastore.ProtocolID
	rmcResponse.MethodID = datastore.MethodPerpetuateObject
	rmcResponse.CallID = callID

	return rmcResponse, nil
}
