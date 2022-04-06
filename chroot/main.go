package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/google/go-containerregistry/pkg/crane"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	images := os.Args[1:]

	for _, image := range images {
		imageHash, entrypoint := pullAndExtractImage(image)
		cmd := exec.Command(entrypoint)
		cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: fmt.Sprintf("/chroot/%s", imageHash)}
		var inputb, outb, errb bytes.Buffer
		inputb.Write(input)
		cmd.Stdin = &inputb
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			fmt.Printf("Stdout: \n%s\n\n", outb.String())
			fmt.Printf("Stderr: \n%s\n\n", errb.String())
			panic(err)
		}
		fmt.Printf("Stdout: \n%s\n\n", outb.String())
		fmt.Printf("Stderr: \n%s\n\n", errb.String())
		input = outb.Bytes()
	}
}

func pullAndExtractImage(imageName string) (string, string) {
	fmt.Printf("Pulling image %s ...\n", imageName)
	image, err := crane.Pull(imageName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Pulled image %s!\n", imageName)

	fmt.Printf("Extracting image %s ...\n", imageName)
	r := mutate.Extract(image)
	tr := tar.NewReader(r)
	imageHash, err := image.Digest()
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(fmt.Sprintf("/chroot/%s", imageHash), os.ModePerm); err != nil {
		panic(err)
	}
	if err := Untar(fmt.Sprintf("/chroot/%s", imageHash), tr); err != nil {
		panic(err)
	}
	fmt.Printf("Extracted image %s ...\n", imageName)

	_, err = exec.Command("cp", "-r", "/bin", fmt.Sprintf("/chroot/%s/", imageHash)).Output()
	if err != nil {
		panic(err)
	}
	_, err = exec.Command("cp", "-r", "/lib", fmt.Sprintf("/chroot/%s/", imageHash)).Output()
	if err != nil {
		panic(err)
	}

	cf, err := image.ConfigFile()
	if err != nil {
		panic(err)
	}

	return imageHash.String(), cf.Config.Entrypoint[0]
}

func Untar(dst string, tr *tar.Reader) error {
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		target := filepath.Join(dst, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
		}
	}
}
