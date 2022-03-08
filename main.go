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
	policyConfigMediaType = "application/vnd.cncf.kyverno.config.v1+json"
	policyLayerMediaType  = "application/vnd.cncf.kyverno.policy.layer.v1+yaml"
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

	img := mutate.MediaType(empty.Image, types.OCIManifestSchema1).(v1.Image)
	img = mutate.ConfigMediaType(img, policyConfigMediaType).(v1.Image)
	// we might set these annotations as labels in the config
	//  img, err = mutate.ConfigFile(img, &v1.ConfigFile{
	//		OS:           "wasi",
	//		Architecture: "wasm",
	//		Variant:      *variant,
	//		Config: v1.Config{
	//			Entrypoint: []string{fn},
	//		},
	//	})
	img = mutate.Annotations(img, annotations).(v1.Image)
	policyLayer := static.NewLayer(policyBytes, policyLayerMediaType)

	img, err = mutate.Append(img, mutate.Addendum{
		Layer: policyLayer,
	})

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Uploading Kyverno policy file [%s] to [%s] with mediaType [%s].\n", policyRef, imageRef.Name(), policyLayerMediaType)
	if err = remote.Write(imageRef, img, defaultOptions...); err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Kyverno policy file [%s] successfully uploaded to [%s]\n", policyRef, imageRef.Name())
}
