package github_test

import (
	"github.com/onsi/gomega/format"
	"testing"

	_ "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func init() {
	format.MaxLength = 8000
}

func TestBackend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitHub Suite")
}

//var _ = BeforeSuite(func() {
//	dbRunner = db.NewRunner()
//	Expect(dbRunner.Start()).To(Succeed())
//
//	dbClient = db.NewClient()
//	Expect(dbClient.Connect(dbRunner.Address())).To(Succeed())
//})
//
//var _ = AfterSuite(func() {
//	Expect(dbRunner.Stop()).To(Succeed())
//})
