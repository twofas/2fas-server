package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
}

func (suite *IntegrationTestSuite) SetupTest() {

}

func Test_IntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
