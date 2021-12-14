package elemental

import (
	"fmt"
	. "github.com/onsi/gomega"
	v1 "github.com/rancher-sandbox/elemental-cli/pkg/types/v1"
	v1mock "github.com/rancher-sandbox/elemental-cli/tests/mocks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"io/ioutil"
	"os"
	"testing"
)

func TestDoCopyEmpty(t *testing.T) {
	RegisterTestingT(t)
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	s, err := os.MkdirTemp("", "elemental")
	Expect(err).To(BeNil())
	defer os.RemoveAll(s)
	d, err := os.MkdirTemp("", "elemental")
	Expect(err).To(BeNil())
	defer os.RemoveAll(d)

	cfg := &v1.RunConfig{
		Device:    "",
		Target:    d,
		Source:    s,
		CloudInit: "",
		Logger:    logger,
	}

	c := Elemental{config: cfg}

	err = c.CopyCos()
	Expect(err).To(BeNil())
}

func TestDoCopy(t *testing.T) {
	RegisterTestingT(t)
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	s, err := os.MkdirTemp("", "elemental")
	Expect(err).To(BeNil())
	defer os.RemoveAll(s)
	d, err := os.MkdirTemp("", "elemental")
	Expect(err).To(BeNil())
	defer os.RemoveAll(d)

	for i := 0; i < 5; i++ {
		_, _ = os.CreateTemp(s, "file*")
	}

	cfg := &v1.RunConfig{
		Device:    "",
		Target:    d,
		Source:    s,
		CloudInit: "",
		Logger:    logger,
	}

	c := Elemental{config: cfg}
	err = c.CopyCos()
	Expect(err).To(BeNil())

	filesDest, err := ioutil.ReadDir(d)
	destNames := getNamesFromListFiles(filesDest)
	filesSource, err := ioutil.ReadDir(s)
	SourceNames := getNamesFromListFiles(filesSource)

	// Should be the same files in both dirs now
	Expect(destNames).To(Equal(SourceNames))
}

func TestDoCopyEmptyWithCloudInit(t *testing.T) {
	RegisterTestingT(t)
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	testString := "In a galaxy far far away..."
	s, err := os.MkdirTemp("", "elemental")
	Expect(err).To(BeNil())
	defer os.RemoveAll(s)
	d, err := os.MkdirTemp("", "elemental")
	Expect(err).To(BeNil())
	defer os.RemoveAll(d)
	err = os.Mkdir(fmt.Sprintf("%s/oem", d), 0777)
	Expect(err).To(BeNil())

	cloudInit, err := os.CreateTemp("", "elemental*")
	_, err = cloudInit.WriteString(testString)
	Expect(err).To(BeNil())
	err = cloudInit.Close()
	Expect(err).To(BeNil())
	defer os.Remove(cloudInit.Name())

	cfg := &v1.RunConfig{
		Target:    d,
		Source:    s,
		CloudInit: cloudInit.Name(),
		Logger:    logger,
	}

	c := Elemental{config: cfg}
	err = c.CopyCos()
	Expect(err).To(BeNil())
	err = c.CopyCloudConfig()
	Expect(err).To(BeNil())
	filesDest, err := ioutil.ReadDir(fmt.Sprintf("%s/oem", d))
	destNames := getNamesFromListFiles(filesDest)

	Expect(destNames).To(ContainElement("99_custom.yaml"))

	dest, err := ioutil.ReadFile(fmt.Sprintf("%s/oem/99_custom.yaml", d))
	Expect(dest).To(ContainSubstring(testString))

}

func TestSelinuxRelabel(t *testing.T) {
	// I cant seem to mock exec.LookPath so it will always fail tor un due setfiles not being in the system :/
	RegisterTestingT(t)
	fs := afero.NewMemMapFs()
	cfg := &v1.RunConfig{Target: "/", Fs: fs}
	c := Elemental{config: cfg}
	// This is actually failing but not sure we should return an error
	Expect(c.SelinuxRelabel(true)).ToNot(BeNil())
	fs = afero.NewMemMapFs()
	_, _ = fs.Create("/etc/selinux/targeted/contexts/files/file_contexts")
	Expect(c.SelinuxRelabel(false)).To(BeNil())
}

func TestCheckFormat(t *testing.T) {
	RegisterTestingT(t)
	fs := afero.NewMemMapFs()
	cfg := &v1.RunConfig{Target: "/", Fs: fs}
	cos := NewElemental(cfg)
	err := cos.CheckNoFormat()
	Expect(err).To(BeNil())
}

func TestCheckNoFormat(t *testing.T) {
	RegisterTestingT(t)
	fs := afero.NewMemMapFs()
	runner := v1mock.FakeRunner{}
	cfg := &v1.RunConfig{Target: "/", Fs: fs, NoFormat: true, Runner: &runner}
	cos := NewElemental(cfg)
	err := cos.CheckNoFormat()
	Expect(err).To(BeNil())
}

// TestCheckNoFormatWithLabel tests when we set no format but labels exists for active/passive partition
func TestCheckNoFormatWithLabel(t *testing.T) {
	RegisterTestingT(t)
	fs := afero.NewMemMapFs()
	logger := v1.NewNullLogger()
	runner := v1mock.NewTestRunnerV2()
	runner.ReturnValue = []byte("/dev/fake")
	cfg := &v1.RunConfig{Target: "/", Fs: fs, NoFormat: true, Runner: runner, Logger: logger}
	cos := NewElemental(cfg)
	err := cos.CheckNoFormat()
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring("There is already an active deployment"))
}

// TestCheckNoFormatWithLabel tests when we set no format but labels exists for active/passive partition AND we set the force flag
func TestCheckNoFormatWithLabelAndForce(t *testing.T) {
	RegisterTestingT(t)
	fs := afero.NewMemMapFs()
	logger := v1.NewNullLogger()
	runner := v1mock.NewTestRunnerV2()
	runner.ReturnValue = []byte("/dev/fake")
	cfg := &v1.RunConfig{Target: "/", Fs: fs, NoFormat: true, Force: true, Runner: runner, Logger: logger}
	cos := NewElemental(cfg)
	err := cos.CheckNoFormat()
	Expect(err).To(BeNil())
}

func getNamesFromListFiles(list []os.FileInfo) []string {
	var names []string
	for _, f := range list {
		names = append(names, f.Name())
	}
	return names
}