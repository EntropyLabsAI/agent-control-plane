# Sentinel 

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/asteroidai/sentinel/server)](https://goreportcard.com/report/github.com/asteroidai/sentinel/server)
[![GitHub stars](https://img.shields.io/github/stars/asteroidai/sentinel?style=social)](https://github.com/asteroidai/sentinel/stargazers)
[![PyPI version](https://badge.fury.io/py/asteroid-sdk.svg)](https://badge.fury.io/py/asteroid-sdk)
[![Downloads](https://pepy.tech/badge/asteroid-sdk)](https://pepy.tech/project/asteroid-sdk)

> ❗ **Important Notice:** 
> 
> - **This repository is no longer actively maintained by Asteroid.**
> - You are welcome to use it under the MIT license, but please note:
>   - We are actively developing new updates on the [Asteroid platform](http://asteroid.ai/) and the [Asteroid SDK](https://github.com/asteroidai/asteroid-python-sdk).
>   - For practical examples and guidance on using the SDK, please explore our [cookbook](https://github.com/asteroidai/cookbook).
> - If you're interested in leveraging the full capabilities of Asteroid, we invite you to visit our [website](http://asteroid.ai/) and get in touch with us!



Sentinel is an agent control plane built by [Asteroid](http://asteroid.ai/) that allows you to efficiently oversee thousands of agents running in parallel.

🎉 New: [Inspect](https://inspect.ai-safety-institute.org.uk/) has now made approvals a native feature! Check out the Inspect example [here](examples/inspect_example/README.md).

## Sentinel Demo Video
[![Sentinel Demo Video](thumb.png)](https://www.youtube.com/watch?v=pOfnYkdLk18)

🚀 Want to see Sentinel in action or chat about agent supervision? [Book a demo with us](https://calendly.com/founders-asteroid-hhaf/30min)!

We're starting with manual reviews for agent actions, but we'll add ways to automatically approve known safe actions in the future.

## Getting Started

See our docs for examples of how to use Sentinel with any agent https://docs.asteroid.ai/quickstart

This repo contains a simple web server written in Go and a React frontend. Agent code can make use of our SDK to make requests to our [API](https://docs.asteroid.ai/api-reference/project/get-all-projects) when an agent makes tool calls, which will be visible in the Sentinel UI. 

1. Start the webserver and frontend with docker compose:
```bash
cp .env.example .env # Set the environment variables in the .env file
source .env          # Pick up the environment variables
docker compose up    # Start the server and frontend
```

2. Run an agent that is pointing at Sentinel via our SDK. See the [examples](/examples) for more details.

For more details, see our [docs](https://docs.asteroid.ai/introduction).

## Examples
We have a number of example containing agents that are using the Sentinel SDK. These are ready to try out of the box:
- [Inspect](https://docs.asteroid.ai/inspect)
- [OpenAI](https://docs.asteroid.ai/openai)

## Development

See https://docs.asteroid.ai/development

## Release

```bash
git tag v0.0.1
git push origin v0.0.1
```
