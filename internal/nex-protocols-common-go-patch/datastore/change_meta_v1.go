package datastore

import (
	"github.com/PretendoNetwork/nex-go/v2"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
)

// changeMetaV1 serves the pre-revision-1 ChangeMeta call (method 0x06), which
// the retail Injustice client sends right after a match is made. It is
// ChangeMeta without the compare / persistence-target extras.
func (commonProtocol *CommonProtocol) changeMetaV1(err error, packet nex.PacketInterface, callID uint32, param datastore_types.DataStoreChangeMetaParamV1) (*nex.RMCMessage, *nex.Error) {
	current := datastore_types.NewDataStoreChangeMetaParam()
	current.DataID = param.DataID
	current.ModifiesFlag = param.ModifiesFlag
	current.Name = param.Name
	current.Permission = param.Permission
	current.DelPermission = param.DelPermission
	current.Period = param.Period
	current.MetaBinary = param.MetaBinary
	current.Tags = param.Tags
	current.UpdatePassword = param.UpdatePassword

	response, rmcErr := commonProtocol.changeMeta(err, packet, callID, current)
	if response != nil {
		response.MethodID = datastore.MethodChangeMetaV1
	}

	return response, rmcErr
}
