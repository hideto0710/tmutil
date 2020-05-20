# Torchstand

## Usage
```bash
Utilities for TorchServe

Usage:
  torchstand [command]

Available Commands:
  build       Build a model archive for torchstand
  help        Help about any command
  import      Import a model from mar file
  models      List TorchServe models
  pull        Pull a model from a registry
  push        Push a model to a registry
  rmm         Remove model reference
  run         Run model server in a new container
  save        Save PyTorch Model Archive to a mar archive (streamed to STDOUT)
  tag         Create a tag that refers to source model

Flags:
      --config string   config file (default is $HOME/.torchstand.yaml)
  -h, --help            help for torchstand
      --insecure        allow connections to SSL registry without certs
      --plain-http      use plain http and not https
      --verbose         verbose output
      --version         Show the TorchStand version information
```

### Quick Start
```bash
curl -o example/densenet161-8d451a50.pth https://download.pytorch.org/models/densenet161-8d451a50.pth

# build model
torchstand build -f example/torchserve.yaml -t localhost:5000/densenet161:v1

# list models
torchstand models

# run TorchServer localy
torchstand run localhost:5000/densenet161:v1 -p 8080

# push to registry
torchstand push localhost:5000/densenet161:v1

# untaged model
torchstand rmm localhost:5000/densenet161:v1

# pull from registry
torchstand pull localhost:5000/densenet161:v1
```

```bash
cd example
curl -O https://s3.amazonaws.com/model-server/inputs/kitten.jpg
curl -X POST "http://0.0.0.0:8080/predictions/densenet161" -T kitten.jpg
```
