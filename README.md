# assignment_demo_2023

![Tests](https://github.com/TikTokTechImmersion/assignment_demo_2023/actions/workflows/test.yml/badge.svg)

This is a demo and template for backend assignment of 2023 TikTok Tech Immersion.

From Sihan:
    Sorry for this incomplete assignment. Because I have zero basic knowledge in these topics, the learning curve for this assignment is way too steep for me and I tried to do as much as possible.

## Installation

Requirement:

- golang 1.18+
- docker

To install dependency tools:

```bash
make pre
```

## Run

```bash
docker-compose up -d
```

Check if it's running:

```bash
curl localhost:8080/ping
```
