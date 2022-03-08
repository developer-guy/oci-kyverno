# Prerequisites

* crane 
* docker
* go
* jq

# Usage

```shell
$ docker run -d -p 5000:5000 --restart=always --name registry registry:2

$ go run main.go disallow-capabilities.yaml localhost:5000/disallow-capabilities:latest
Uploading Kyverno policy file [disallow-capabilities.yaml] to [localhost:5000/disallow-capabilities:latest] with mediaType [application/vnd.cncf.kyverno.policy.layer.v1+yaml].
Kyverno policy file [disallow-capabilities.yaml] successfully uploaded to [localhost:5000/disallow-capabilities:latest]

$ crane manifest localhost:5000/disallow-capabilities:latest | jq 
{
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.cncf.kyverno.config.v1+json",
    "size": 233,
    "digest": "sha256:d924710ff69c27353cee743d00226e7b1bd461b6df16943d983738e5264dfb3d"
  },
  "layers": [
    {
      "mediaType": "application/vnd.cncf.kyverno.policy.layer.v1+yaml",
      "size": 1551,
      "digest": "sha256:5b6075facc39bd992695f2c44285ae78165cf1497539b49168da4698a16cbfe7"
    }
  ],
  "annotations": {
    "kyverno.io/kubernetes-version": "1.22-1.23",
    "kyverno.io/kyverno-version": "1.6.0",
    "policies.kyverno.io/category": "Pod Security Standards (Baseline)",
    "policies.kyverno.io/description": "Adding capabilities beyond those listed in the policy must be disallowed.",
    "policies.kyverno.io/minversion": "1.6.0",
    "policies.kyverno.io/severity": "medium",
    "policies.kyverno.io/subject": "Pod",
    "policies.kyverno.io/title": "Disallow Capabilities"
  }
}
```
