# CF Provider Timeline V2

https://cf-timeline.crsimmons.dev

This is a Cloud Foundry application for comparing the public CF providers based on when the version of the [Cloud Controller API](https://github.com/cloudfoundry/cloud_controller_ng) they are running was released.

It does this by pulling capi versions from each provider via `v2/info` on their API endpoints. Then it parses the release text of the [CAPI BOSH release](https://github.com/cloudfoundry/capi-release) looking for the version of the API included in that release.

The front end is inspired from an app I built with @jpluscplusm a few years ago which served the same purpose. It has since stopped working due to some of the websites we were scraping chaging their schema.

## Adding APIs

Do you run a public CF PaaS that isn't included? You can add it by PRing your API to [this file](assets/providers.json).
