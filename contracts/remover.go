package contracts

import "github.com/goodsru/go-universal-network-adapter/models"

type Remover interface {
	Remove(remoteFile *models.RemoteFile) error
}
