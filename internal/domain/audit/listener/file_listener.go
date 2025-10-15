package listener

import (
	"encoding/json"
	"os"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit/data"
)

type FileListener struct {
	path string
}

func NewFileListener(path string) *FileListener {
	return &FileListener{path: path}
}

func (fl *FileListener) Notify(data *data.Data) error {
	existData, err := fl.LoadFile()
	if err != nil {
		return err
	}

	existData = append(existData, *data)

	file, err := os.OpenFile(fl.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	return json.NewEncoder(file).Encode(existData)
}

func (fl *FileListener) LoadFile() (data []data.Data, err error) {
	if _, err = os.Stat(fl.path); err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}

	logs, err := os.ReadFile(fl.path)
	if err != nil {
		return
	}

	if len(logs) > 0 {
		err = json.Unmarshal(logs, &data)
		return
	}
	return
}
