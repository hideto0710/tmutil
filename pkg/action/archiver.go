package action

import (
	"archive/tar"
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	torchstandTypes "github.com/hideto0710/torchstand/pkg/types"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Archiver struct {
	ref          *torchstandTypes.Ref
	registryPath string
}

func NewArchiver(ref *torchstandTypes.Ref, registryPath string) *Archiver {
	return &Archiver{
		ref: ref, registryPath: registryPath,
	}
}

func (a *Archiver) Archive(dest string) error {
	zipFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	w := zip.NewWriter(zipFile)
	defer w.Close()

	if err := a.copyConfig(w); err != nil {
		return err
	}
	if err := a.copyPyTorchModel(w); err != nil {
		return err
	}
	if err := a.copyContents(w); err != nil {
		return err
	}
	return nil
}

func (a *Archiver) copyConfig(w *zip.Writer) error {
	return addFileToZip(w, a.blobPath(a.ref.Config), marFilePath)
}

func (a *Archiver) copyPyTorchModel(w *zip.Writer) error {
	// TODO: read config.json
	return addFileToZip(w, a.blobPath(a.ref.PyTorchModel), "densenet161-8d451a50.pth")
}

func (a *Archiver) copyContents(w *zip.Writer) error {
	contentFile, err := os.Open(a.blobPath(a.ref.Content))
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(contentFile)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := header.Name
		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			f, err := w.Create(name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

func (a *Archiver) blobPath(desc v1.Descriptor) string {
	return filepath.Join(a.registryPath, "blobs", desc.Digest.Algorithm().String(), desc.Digest.Hex())
}
