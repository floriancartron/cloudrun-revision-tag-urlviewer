# Cloud Run Revision Tag URL Viewer

This application fetches data from Google Cloud Run services in order to show URL that can be accessed through Serverless Neg with url masks.

## Features

- Fetches and displays Cloud Run service revision tags
- Displays URLs with Cloud Run revision tags
- Shows deployment date and other metadata for services
- Provides a simple health check endpoint

## Prerequisites

- Your Cloud Run Services should have an annotation with the base url for the service. The annotation should be named `baseurl`. For example, if your service is exposed with a URL mask like `<tag>.example.com`, the annoation should be `baseurl: example.com`.

## Run the app

```shell
docker build -t crrtu .
docker run -p 8080:8080 \
  -v $HOME/.config/gcloud/application_default_credentials.json:/app/application_default_credentials.json \
  -e GOOGLE_APPLICATION_CREDENTIALS=/app/application_default_credentials.json \
  -e CRRTUV_PROJECT="your-project-with-url-masks" \
  -e CRRTUV_LOCATION="your-region" \
  -e CRRTUV_IDENTIFYING_LABEL="an-additional-label-that-appear" \
  -e CRRTUV_TIMEZONE="timezone-to-use" \
  -e CRRTUV_MAX_REVISIONS="max-revisions-to-fetch-for-each-service,defaults to 100" \
   crrtu
```

It should then be available at http://localhost:8080