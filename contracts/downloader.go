// universal-network-adapter interfaces
package contracts

import "github.com/konart/go-universal-network-adapter/models"

type Downloader interface {
	// Get remote file info
	Stat(destination *models.ParsedDestination) (*models.RemoteFile, error)
	// Browse remote directory
	Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error)
	// Download remote file
	Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error)
}
