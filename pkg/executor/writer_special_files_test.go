package executor

import (
	"archive/tar"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	specs "github.com/geaaru/tar-formers/pkg/specs"
	tools "github.com/geaaru/tar-formers/pkg/tools"
)

// A FIFO must be archived as a FIFO header without opening it for reading:
// open(2) on a FIFO with no writer blocks forever. Sockets cannot be stored
// in a tar archive at all and must be skipped instead of failing the run.
func TestInjectDirWithFifoAndSocket(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "regular"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(src, "fifo"), 0640); err != nil {
		t.Fatal(err)
	}

	socketPath := filepath.Join(src, "socket")
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	out := filepath.Join(t.TempDir(), "out.tar")

	cfg := specs.NewConfig(nil)
	cfg.GetLogging().Level = "error"
	tarformers := NewTarFormers(cfg)

	tarfspec := specs.NewSpecFile()
	tarfspec.Writer = specs.NewWriter()
	tarfspec.Writer.AddDir(src)
	tarfspec.RenamePath = append(tarfspec.RenamePath,
		specs.RenameRule{Source: src, Dest: "."})

	opts := tools.NewTarCompressionOpts(true)
	defer opts.Close()
	if err := tools.PrepareTarWriter(out, opts); err != nil {
		t.Fatal(err)
	}
	if opts.CompressWriter != nil {
		tarformers.SetWriter(opts.CompressWriter)
	} else {
		tarformers.SetWriter(opts.FileWriter)
	}

	done := make(chan error, 1)
	go func() {
		done <- tarformers.RunTaskWriter(tarfspec)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunTaskWriter failed: %s", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("RunTaskWriter is blocked on a special file")
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tr := tar.NewReader(io.Reader(f))
	got := map[string]byte{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got[filepath.Base(hdr.Name)] = hdr.Typeflag
	}

	if got["fifo"] != tar.TypeFifo {
		t.Fatalf("fifo entry missing or wrong type: %q", got["fifo"])
	}
	if _, ok := got["socket"]; ok {
		t.Fatal("socket entry must be skipped")
	}
}
