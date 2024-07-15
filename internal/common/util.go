package common

import (
	"github.com/go-logr/logr"
	mfc "github.com/manifestival/controller-runtime-client"
	"github.com/manifestival/manifestival"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetupManifestival instantiates a Manifestival instance for the provided file or directory
func SetupManifestival(client client.Client, fileOrDir string, recurse bool, logger logr.Logger) (manifestival.Manifest, error) {
	mfclient := mfc.NewClient(client)

	// manifest := filepath.Join(dataPath, fileOrDir)
	var src manifestival.Source
	if recurse {
		src = manifestival.Recursive(fileOrDir)
	} else {
		src = manifestival.Path(fileOrDir)
	}

	return manifestival.ManifestFrom(src, manifestival.UseClient(mfclient), manifestival.UseLogger(logger))
}
