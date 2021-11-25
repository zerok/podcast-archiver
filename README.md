# Podcast Archiver

This is a tiny utility I wrote to keep a backup copy of every episode of my
favorite shows before they disappear from the internet. Initially, it only
supported S3 compatible backends but was later extended to also store podcasts
to the local filesystem and Dropbox.

## Usage

```sh
./podcast-archiver --config config-file.yaml
```

## Configuration

Podcast Archiver can be configured with a configuration file written in YAML.
It basically sets information about what feeds should be downloaded, into
what folder, and where everything should be stored (the "sink").

The example below would download all episodes it can find in the Changelog
master feed and download them into `./data/changelog-master`.

```yaml
sink:
- filesystem_folder: "./data"
feeds:
- folder: "changelog-master"
  url: "https://changelog.com/master/feed"
```

The following fields are available for the sink:

- `google_project_id`
- `bucket`
- `filesystem_folder`
- `access_key_id`
- `access_key_secret`
- `region`

### Google Cloud Storage sink

If you want to upload the podcasts to a Google Cloud Storage bucket, you'd
need to set the following fields:

- `google_project_id`
- `bucket`

The credentials are taken from the environment using Google Application
Credentials as documented [here][gac].

[gac]: https://developers.google.com/identity/protocols/application-default-credentials

### S3 sink

Here the following fields must be set:

- `bucket`
- `access_key_id`
- `access_key_secret`
- `region`

## Notifications

If you want to get notified whenever a new podcast has been archived, you can
set the following environment variables to send a message to a pre-defined
Matrix room:

- `MATRIX_HOMESERVER`: URL of the homeserver the user is registered at
- `MATRIX_USERNAME`
- `MATRIX_PASSWORD`
- `MATRIX_ROOM`: The complete room name (`!...@...`)
