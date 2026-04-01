# Atlas Migration

## Generate Migration Hash

```bash
docker run --rm -v $(pwd)/deployment/atlas/migrations:/migrations arigaio/atlas:latest migrate hash --dir "file:///migrations"
```