package main

import (
	"errors"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

const (
	kyvernoConfigMediaType      = "application/vnd.cncf.kyverno.config.v1+json"
	kyvernoPolicyLayerMediaType = "application/vnd.cncf.kyverno.policy.layer.v1+yaml"
)

func main() {
	if len(os.Args) < 1 {
		panic(errors.New("you should specify policy path as a first argument and an image as a second argument"))
	}

	policyRef := os.Args[1]
	image := os.Args[2]

	policyBytes, err := os.ReadFile(filepath.Clean(policyRef))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var policyMap map[string]interface{}
	err = yaml.Unmarshal(policyBytes, &policyMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	annotations := map[string]string{}
	for k, v := range policyMap["metadata"].(map[string]interface{})["annotations"].(map[string]interface{}) {
		annotations[k] = v.(string)
	}

	imageRef, err := name.ParseReference(image)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var defaultOptions = []remote.Option{
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	fmt.Fprintf(os.Stderr, "Uploading Kyverno policy file [%s] to [%s] with mediaType [%s].\n", policyRef, imageRef.Name(), kyvernoPolicyLayerMediaType)
	base := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	base = mutate.ConfigMediaType(base, kyvernoConfigMediaType)
	base = mutate.Annotations(base, annotations).(v1.Image)
	policyLayer := static.NewLayer(policyBytes, kyvernoPolicyLayerMediaType)

	img, err := mutate.Append(base, mutate.Addendum{
		Layer: policyLayer,
	})

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = remote.Write(imageRef, img, defaultOptions...)

	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "Kyverno policy file [%s] successfully uploaded to [%s]\n", policyRef, imageRef.Name())
}
