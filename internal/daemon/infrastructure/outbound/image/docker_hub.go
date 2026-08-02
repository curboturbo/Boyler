package image

import (
	"boyler/internal/daemon/core"
	logger "boyler/pkg/logger"
	pkg_string "boyler/pkg/string"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const layersInfoFileName = "layers.json"

type DockerHubPuller struct {
	HTTPClient *http.Client
	Platform   Platform
	Progress chan *core.PullingEvent
}


type Platform struct {
	OS           string
	Architecture string
}


func NewDockerHubPuller(osSettings Platform, ch chan *core.PullingEvent) *DockerHubPuller {
	return &DockerHubPuller{
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		Platform:   Platform{OS: osSettings.OS, Architecture: osSettings.Architecture},
		Progress: ch,
	}
}


func (p *DockerHubPuller) Supports(ref string) bool {
	return !strings.Contains(ref, "/") || strings.HasPrefix(ref, "docker.io/") || !strings.Contains(strings.Split(ref, "/")[0], ".")
}

func (p *DockerHubPuller) dockerHubToken(ctx context.Context, repository string) (string, error) {
	log := logger.FromContext(ctx)
	log.Debug("Start request dockerHub token","repo",repository)
	if !strings.Contains(repository, "/") {
		repository = "library/" + repository
	}

	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repository)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth token endpoint returned %s", resp.Status)
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	log.Debug("Received auth token", "token",body.Token[:10])
	return body.Token, nil
}


func (p *DockerHubPuller) Pull(ctx context.Context, ref string, destDir string) (string, error) {
	log := logger.FromContext(ctx)
	log.Debug("Start pulling image","image",ref)
	ref = strings.TrimPrefix(ref, "docker.io/")
	parts := strings.SplitN(ref, ":", 2)
	repo := parts[0]
	tag := "latest"
	if len(parts) == 2 {
		tag = parts[1]
	}

	if !strings.Contains(repo, "/") {
		repo = "library/" + repo
	}

	token, err := p.dockerHubToken(ctx, repo)
	if err != nil {
		return "", fmt.Errorf("get token: %w", err)
	}

	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", repo, tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.index.v1+json, application/vnd.oci.image.manifest.v1+json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<10))
		return "", fmt.Errorf("manifest fetch returned %s: %s", resp.Status, string(b))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	digest := resp.Header.Get("Docker-Content-Digest")

	ct := resp.Header.Get("Content-Type")
	var manifest ociManifest
	if strings.Contains(ct, "manifest.list.v2+json") || strings.Contains(ct, "image.index.v1+json") {
		var idx ociIndex
		if err := json.Unmarshal(bodyBytes, &idx); err != nil {
			return "", err
		}
		var subDigest string
		for _, m := range idx.Manifests {
			if m.Platform.OS == p.Platform.OS && m.Platform.Architecture == p.Platform.Architecture {
				subDigest = m.Digest
				break
			}
		}
		if subDigest == "" {
			return "", fmt.Errorf("no manifest for platform %s/%s", p.Platform.OS, p.Platform.Architecture)
		}

		subURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", repo, subDigest)
		subReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, subURL, nil)
		subReq.Header.Set("Authorization", "Bearer "+token)
		subReq.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
		subResp, err := p.HTTPClient.Do(subReq)
		if err != nil {
			return "", err
		}
		defer subResp.Body.Close()
		if err := json.NewDecoder(subResp.Body).Decode(&manifest); err != nil {
			return "", err
		}
	} else {
		if err := json.Unmarshal(bodyBytes, &manifest); err != nil {
			return "", err
		}
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	safeName := pkg_string.SanitizeImageName(ref)
	imagePath := filepath.Join(destDir, safeName)
	if err := p.isExistOr(imagePath); err != nil {
		return "", err
	}
	num := 0
	for i, layer := range manifest.Layers {
		num +=1
		blobURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/blobs/%s",repo,layer.Digest)
		bReq, err := http.NewRequestWithContext(ctx,http.MethodGet,blobURL,nil,)
		if err != nil {
			return "", err
		}
		bReq.Header.Set("Authorization","Bearer "+token)
		bResp, err := p.HTTPClient.Do(bReq)
		if err != nil {
			return "", err
		}

		if bResp.StatusCode != http.StatusOK {
			bResp.Body.Close()
			return "", fmt.Errorf("download layer %s failed: %s",layer.Digest,bResp.Status)
		}
		fileName := fmt.Sprintf("layer_%d.tar.gz",i)
		outPath := filepath.Join(imagePath, fileName)
		outFile, err := os.Create(outPath)
		if err != nil {
			bResp.Body.Close()
			return "", err
		}
		total := bResp.ContentLength
		var downloaded int64
		buffer := make([]byte, 1024*1024)
		for {
			n, err := bResp.Body.Read(buffer)
			if n > 0 {
				_, writeErr := outFile.Write(buffer[:n])
				if writeErr != nil {
					outFile.Close()
					bResp.Body.Close()
					return "", writeErr
				}
			downloaded += int64(n)

			p.Progress <- &core.PullingEvent{
					LayId: layer.Digest[:12],
					Status: "Downloading",
					Progress: downloaded,
					Total: total,
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				outFile.Close()
				bResp.Body.Close()
				return "", fmt.Errorf("read layer %s: %w",layer.Digest,err)
			}
		}
		err = outFile.Close()
		bResp.Body.Close()

		if err != nil {
			return "", err
		}

		p.Progress <- &core.PullingEvent{
				LayId: layer.Digest[:12],
				Status: "Pull complete",
				Progress: total,
				Total: total,
			}
	}
	createLayersInfo(num,imagePath)
	return digest, nil
}

func (p *DockerHubPuller) isExistOr(path string) error {
    if _, err := os.Stat(path); err == nil {
        return fmt.Errorf("image has already been downloaded")
    } else if !os.IsNotExist(err) {
        return err
    }
    if err := os.MkdirAll(path, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    return nil
}


func createLayersInfo(num int, saveDir string) error {
	info := layersInfo{Num: num}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal layers info: %w", err)
	}
	path := filepath.Join(saveDir, layersInfoFileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write layers info: %w", err)
	}

	return nil
}

type ociDescriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type ociManifest struct {
	SchemaVersion int             `json:"schemaVersion"`
	Layers        []ociDescriptor `json:"layers"`
}

type ociIndexEntry struct {
	ociDescriptor
	Platform struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
	} `json:"platform"`
}

type ociIndex struct {
	Manifests []ociIndexEntry `json:"manifests"`
}

type layersInfo struct {
	Num int `json:"num"`
}

