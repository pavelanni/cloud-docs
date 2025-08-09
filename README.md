# Cloud Docs

This project aims at hosting documents on cloud-native object storage and serving them in a serverless manner.
It also implements basic access control (currently via a simple token) to hide the docs from public eye and search engines.

## Plan

1. The application should be developed to run on Google Cloud Platform.
1. The application will use Google Cloud Storage to host HTML, CSS, JS, and other resources.
1. The application will use Google Cloud Run for serverless deployment.
1. The application will be written in Go.
1. The application will use a token as part of the the URL to access HTML files and other resources.
1. The project should also include a tool to upload HTML and other resources to the cloud storage while preserving directory structure.
1. The project should also include a tool to produce an `iframe` element with a specified document path. That iframe element should include the URL with the token.
1. The project should also include a tool to create a token that will be included in the document URLs and which will be checked by the middleware part of the Go-based web server.


