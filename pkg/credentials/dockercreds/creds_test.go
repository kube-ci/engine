package dockercreds

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"kube.ci/kubeci/pkg/credentials"
)

func TestFlagHandling(t *testing.T) {
	credentials.VolumePath, _ = ioutil.TempDir("", "")
	dir := credentials.VolumeName("foo")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", dir, err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, corev1.BasicAuthUsernameKey), []byte("bar"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(username) = %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, corev1.BasicAuthPasswordKey), []byte("baz"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(password) = %v", err)
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags(fs)
	err := fs.Parse([]string{
		"-basic-docker=foo=https://us.gcr.io",
	})
	if err != nil {
		t.Fatalf("flag.CommandLine.Parse() = %v", err)
	}

	os.Setenv("HOME", credentials.VolumePath)
	if err := NewBuilder().Write(); err != nil {
		t.Fatalf("Write() = %v", err)
	}

	b, err := ioutil.ReadFile(filepath.Join(credentials.VolumePath, ".docker", "config.json"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile(.docker/config.json) = %v", err)
	}

	// Note: "auth" is base64(username + ":" + password)
	expected := `{"auths":{"https://us.gcr.io":{"username":"bar","password":"baz","auth":"YmFyOmJheg==","email":"not@val.id"}}}`
	if string(b) != expected {
		t.Errorf("got: %v, wanted: %v", string(b), expected)
	}
}

func TestFlagHandlingTwice(t *testing.T) {
	credentials.VolumePath, _ = ioutil.TempDir("", "")
	fooDir := credentials.VolumeName("foo")
	if err := os.MkdirAll(fooDir, os.ModePerm); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", fooDir, err)
	}
	if err := ioutil.WriteFile(filepath.Join(fooDir, corev1.BasicAuthUsernameKey), []byte("asdf"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(username) = %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(fooDir, corev1.BasicAuthPasswordKey), []byte("blah"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(password) = %v", err)
	}
	barDir := credentials.VolumeName("bar")
	if err := os.MkdirAll(barDir, os.ModePerm); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", barDir, err)
	}
	if err := ioutil.WriteFile(filepath.Join(barDir, corev1.BasicAuthUsernameKey), []byte("bleh"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(username) = %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(barDir, corev1.BasicAuthPasswordKey), []byte("belch"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(password) = %v", err)
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags(fs)
	err := fs.Parse([]string{
		"-basic-docker=foo=https://us.gcr.io",
		"-basic-docker=bar=https://eu.gcr.io",
	})
	if err != nil {
		t.Fatalf("flag.CommandLine.Parse() = %v", err)
	}

	os.Setenv("HOME", credentials.VolumePath)
	if err := NewBuilder().Write(); err != nil {
		t.Fatalf("Write() = %v", err)
	}

	b, err := ioutil.ReadFile(filepath.Join(credentials.VolumePath, ".docker", "config.json"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile(.docker/config.json) = %v", err)
	}

	// Note: "auth" is base64(username + ":" + password)
	expected := `{"auths":{"https://eu.gcr.io":{"username":"bleh","password":"belch","auth":"YmxlaDpiZWxjaA==","email":"not@val.id"},"https://us.gcr.io":{"username":"asdf","password":"blah","auth":"YXNkZjpibGFo","email":"not@val.id"}}}`
	if string(b) != expected {
		t.Errorf("got: %v, wanted: %v", string(b), expected)
	}
}

func TestFlagHandlingMissingFiles(t *testing.T) {
	credentials.VolumePath, _ = ioutil.TempDir("", "")
	dir := credentials.VolumeName("not-found")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", dir, err)
	}
	// No username / password files yields an error.

	cfg := dockerConfig{make(map[string]entry)}
	if err := cfg.Set("not-found=https://us.gcr.io"); err == nil {
		t.Error("Set(); got success, wanted error.")
	}
}

func TestFlagHandlingURLCollision(t *testing.T) {
	credentials.VolumePath, _ = ioutil.TempDir("", "")
	dir := credentials.VolumeName("foo")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", dir, err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, corev1.BasicAuthUsernameKey), []byte("bar"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(username) = %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, corev1.BasicAuthPasswordKey), []byte("baz"), 0777); err != nil {
		t.Fatalf("ioutil.WriteFile(password) = %v", err)
	}

	cfg := dockerConfig{make(map[string]entry)}
	if err := cfg.Set("foo=https://us.gcr.io"); err != nil {
		t.Fatalf("First Set() = %v", err)
	}
	if err := cfg.Set("bar=https://us.gcr.io"); err == nil {
		t.Error("Second Set(); got success, wanted error.")
	}
}

func TestMalformedValueTooMany(t *testing.T) {
	cfg := dockerConfig{make(map[string]entry)}
	if err := cfg.Set("bar=baz=blah"); err == nil {
		t.Error("Second Set(); got success, wanted error.")
	}
}

func TestMalformedValueTooFew(t *testing.T) {
	cfg := dockerConfig{make(map[string]entry)}
	if err := cfg.Set("bar"); err == nil {
		t.Error("Second Set(); got success, wanted error.")
	}
}
