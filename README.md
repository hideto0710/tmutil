# Torchstand

## Usage
```bash
curl -o example/densenet161-8d451a50.pth https://download.pytorch.org/models/densenet161-8d451a50.pth

# archive model
torchstand archive -t localhost:5000/densenet161:v1 -f example/torchserve.yaml

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
