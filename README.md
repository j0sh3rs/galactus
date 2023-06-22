# Galactus: Devourer of Worlds

## Overview

Galactus is a famous Marvel cosmic entity known as the Devourer of Worlds. He consumes entire planets and galaxies to sustain his own life force, and works to ensure order remains across the galaxies.

## What it do

Galactus' purpose is to clean up AMIs and image families (planned) to retire old images from service. It does this by checking an account for AMIs matching a specific pattern (in a specific region) and validates that no instances have launched using this AMI in the last 90 days. It will then devour (deregister) the AMI or instance family so it cannot be used again.

If Galactus finds an image is still in use within the last 90 days, it will leave the image alone.

## Roadmap

- [ ] Unit Tests/Mocks
- [ ] Customizable age for retiring images
- [ ] GCP Support for retiring instance families
- [ ] Add a daemon mode (as default behavior) to use as a k8s service
  - [ ] Cron schedule definition via config/envvar
- [ ] Prometheus Metrics
  - [ ] Number of total culled images
  - [ ] Last runtime epoch
  - [ ] Images pending cleanup (last launched >60d & <90d )
- [ ] Tag images with the last used date for easy referencing/querying
- [ ] Goroutines for watching regions concurrently
  - [ ] Allow for different crons and search strings per region
