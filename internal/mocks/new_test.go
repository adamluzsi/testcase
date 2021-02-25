package mocks_test

import (
	"testing"

	"github.com/adamluzsi/testcase/internal/mocks"
	"github.com/golang/mock/gomock"
)

func TestNew(t *testing.T) {
	_, _ = mocks.New(t)
	_ = mocks.NewMock(t)
	_ = mocks.NewMockTB(gomock.NewController(t))
	_ = mocks.NewWithDefaults(t, func(mock *mocks.MockTB) {})
}
