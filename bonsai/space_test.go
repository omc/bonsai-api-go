package bonsai_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SpaceTestSuite struct {
	*ClientTestSuite
}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(SpaceTestSuite))
}
