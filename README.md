<p align="center">
    <img src="https://user-images.githubusercontent.com/8983173/69229423-ac731300-0b85-11ea-8c2e-2cc00ecdb269.PNG" alt="Gotenberg logo" width="250" height="250" />
</p>
<h3 align="center">Gotenberg</h3>
<p align="center">A Docker-powered stateless API for converting HTML, Markdown and Office documents to PDF.</p>
<p align="center"><a href="https://thecodingmachine.github.io/gotenberg">Documentation</a> &#183; <a href="/.github/CONTRIBUTING.md">Contributing</a></p>

---

At TheCodingMachine, we build a lot of web applications (intranets, extranets and so on) which require to generate PDF from various sources. Each time, we ended up using some well known libraries and kind of lost time by reimplementing a solution from a project to another project. Meh.

## Features

* HTML and Markdown conversions using Google Chrome headless
* Office conversions (.txt, .rtf, .docx, .doc, .odt, .pptx, .ppt, .odp and so on) using [unoconv](https://github.com/dagwieers/unoconv)
* Assets :package:: send your header, footer, images, fonts, stylesheets and so on for converting your HTML and Markdown to beautiful PDFs!
* Easily interact with the API using our [Go](https://github.com/thecodingmachine/gotenberg-go-client) and [PHP](https://github.com/thecodingmachine/gotenberg-php-client) libraries

## Quick start

Open a terminal and run the following command:

```bash
$ docker run --rm -p 3000:3000 thecodingmachine/gotenberg:6
```

The API is now available on your host at `http://localhost:3000`.

Head to the [documentation](https://thecodingmachine.github.io/gotenberg)
to learn how to interact with it!

## Ennova scripts

Run these commands from the repository root:

| Command | Purpose |
| --- | --- |
| `./script/run.sh` | Build and run locally at `http://localhost:3090`. |
| `./script/publish.sh` | Build AMD64 and ARM64 images and publish release tags, version aliases, and `latest` to Docker Hub (`e7build/gotenberg`). Also push `master` and the Git release tag. |
| `./script/publish-ci.sh` | Sync up to 10 published Docker Hub tags to CI's AWS ECR registry, preserving all architectures. |
| `./script/deploy.sh <qa\|prod> [tag]` | Deploy to QA or PROD. Omit the tag to use the latest published release, or provide a release tag explicitly. Run once per environment to deploy to both. |

## Badges

[![Docker pulls](https://img.shields.io/docker/pulls/thecodingmachine/gotenberg)](https://hub.docker.com/r/thecodingmachine/gotenberg)
[![Docker image layers](https://images.microbadger.com/badges/image/thecodingmachine/gotenberg:6.svg)](https://microbadger.com/images/thecodingmachine/gotenberg:6)
[![GitHub Actions](https://github.com/thecodingmachine/gotenberg/workflows/tests_maybe_release/badge.svg)](https://github.com/thecodingmachine/gotenberg/workflows/tests_maybe_release)
[![GoDoc](https://godoc.org/github.com/thecodingmachine/gotenberg?status.svg)](https://godoc.org/github.com/thecodingmachine/gotenberg)
[![Codecov](https://codecov.io/gh/thecodingmachine/gotenberg/branch/master/graph/badge.svg)](https://codecov.io/gh/thecodingmachine/gotenberg)
[![Go Report Card](https://goreportcard.com/badge/github.com/thecodingmachine/gotenberg)](https://goreportcard.com/report/thecodingmachine/gotenberg)

---

<p align="center">
    <img src="https://user-images.githubusercontent.com/8983173/50009948-84b01e00-ffb8-11e8-850b-fc240382c626.png" alt="Gotenberg logo" width="150" height="150" />
</p>
