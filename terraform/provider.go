package terraform

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var providerDefaults = map[string]*Provider{
	"aws": &Provider{
		Name:           "aws",
		Version:        "v4.29.0",
		RepositoryBase: "https://github.com/hashicorp",
		RepositoryName: "terraform-provider-aws",
		DocsBasePath:   "website/docs/",
	},
}

type Provider struct {
	Name           string
	Version        string
	RepositoryBase string
	RepositoryName string
	DocsBasePath   string
	rootDocPath string // e.g. .terrawrap/provider_docs/aws/v4.29.0
}

func (p *Provider) SetRootDocPath(docPath string) {
	p.rootDocPath = docPath
}

func (p *Provider) DocPath() string {
	// TODO: this is specific to AWS provider
	re := regexp.MustCompile(`^v`)
	return path.Join(p.rootDocPath, strings.Join([]string{p.RepositoryName, re.ReplaceAllString(p.Version, "")}, "-"), p.DocsBasePath)
}

func (p *Provider) DownloadDocs() error {
	fileName := p.Version + ".zip"

	tmpFilePath, err := p.downloadFile(fileName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := p.unzipDocFile(tmpFilePath); err != nil {
		return fmt.Errorf("unable to unzip downloaded docs file: %w", err)
	}

	if err := p.cleanUpDownload(tmpFilePath); err != nil {
		return fmt.Errorf("while downloading provider docs: %w", err)
	}

	return err
}

func (p *Provider) unzipDocFile(zipFilePath string) error {
	var (
		err error
		zr  *zip.ReadCloser
		out *os.File
		fr  io.ReadCloser
	)

	zr, err = zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", zipFilePath, err)
	}
	// TODO: checkerr
	defer zr.Close()

	for _, f := range zr.File {
		p := filepath.Join(p.rootDocPath, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", p, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directories for file %s: %w", p, err)
		}

		out, err = os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", p, err)
		}

		fr, err = f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip %s: %w", f.Name, err)
		}

		if _, err := io.Copy(out, fr); err != nil {
			return fmt.Errorf("failed to copy file from zip (%s) to destination (%s): %w", f.Name, p, err)
		}

		if err := out.Close(); err != nil {
			return fmt.Errorf("encountered error while closing output file %s: %w", p, err)
		}

		if err := fr.Close(); err != nil {
			return fmt.Errorf("encountered error while closing archive reader %s: %w", f.Name, err)
		}
	}

	return err
}

func (p *Provider) cleanUpDownload(fileName string) error {
	log.Println("removing downloaded zip file")
	if err := os.Remove(fileName); err != nil {
		return fmt.Errorf("failed to remove downloaded file %s: %w", fileName, err)
	}
	return nil
}

func (p *Provider) downloadFile(fileName string) (string, error) {
	var (
		downloadURI *url.URL
		err         error
		response    *http.Response
		out         *os.File
	)

	targetFilePath := path.Join(os.TempDir(), fileName)
	targetFileDir := filepath.Dir(targetFilePath)
	if _, err := os.Stat(targetFileDir); os.IsNotExist(err) {
		// create directory
		if err := os.MkdirAll(targetFileDir, 0754); err != nil {
			return targetFilePath, fmt.Errorf("failed to create target download directory %s: %w", targetFileDir, err)
		}
	}

	// https://github.com/hashicorp/terraform-provider-aws/archive/refs/tags/v4.29.0.zip
	// Hackish for now
	downloadURI, err = url.Parse(p.RepositoryBase)
	if err != nil {
		return targetFilePath, fmt.Errorf("failed to parse base repository uri %s for provider", p.RepositoryBase)
	}
	downloadURI.Path = path.Join(downloadURI.Path, p.RepositoryName, "archive", "refs", "tags", fileName)

	response, err = http.Get(downloadURI.String())
	if err != nil {
		return targetFilePath, fmt.Errorf("failed to fetch provider from %s: %w", downloadURI.String(), err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return targetFilePath, fmt.Errorf("received status code %d while fetching docs", response.StatusCode)
	}

	out, err = os.Create(targetFilePath)
	if err != nil {
		return targetFilePath, fmt.Errorf("failed to create target docs file %s: %w", targetFilePath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return targetFilePath, fmt.Errorf("failed to write http response to file: %w", err)
	}

	return targetFilePath, nil
}

func GetProvider(name string) (*Provider, error) {
	provider, ok := providerDefaults[name]
	if !ok {
		return nil, fmt.Errorf("failed to locate support for provider %s", name)
	}

	return provider, nil
}
