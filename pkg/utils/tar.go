package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
)

func ExtractTgzBytes(tgzBytes []byte) (map[string][]byte, error) {
	r := bytes.NewReader(tgzBytes)
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gr)

	result := map[string][]byte{}
	for {
		h, err := tr.Next()

		if err == io.EOF {
			return result, nil
		} else if err != nil {
			return result, err
		}

		switch h.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, tr)
			if err != nil {
				return result, err
			}
			result[h.Name] = buf.Bytes()
		default:
			continue
		}
	}
}
