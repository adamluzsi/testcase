package mocks_test

import (
	"github.com/adamluzsi/testcase/internal/mocks"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestNew(t *testing.T) {
	_, _ = mocks.New(t)
	_ = mocks.NewMock(t)
	_ = mocks.NewMockTB(gomock.NewController(t))
	_ = mocks.NewWithDefaults(t, func(mock *mocks.MockTB) {})
}
