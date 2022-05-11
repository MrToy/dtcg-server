package service

import (
	"encoding/json"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type DMPState[T any] struct {
	preStateStr string

	PreState T
	State    T
}

func (s *DMPState[T]) Update(data string) error {
	dmp := diffmatchpatch.New()
	patchs, _ := dmp.PatchFromText(data)
	str, _ := dmp.PatchApply(patchs, s.preStateStr)
	s.preStateStr = str
	s.PreState = s.State
	var state T
	err := json.Unmarshal([]byte(str), &state)
	if err != nil {
		return err
	}
	s.State = state
	return nil
}
