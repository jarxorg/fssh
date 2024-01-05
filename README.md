# fssh

fssh is a shell for multiple file systems.

## Feature

- Multiple file systems
  - local
  - amazon s3
  - google cloud storage
- Command history
- Simple auto complete

## Install

```sh
go install github.com/jarxorg/fssh/cmd/fssh@latest
```

## Examples

### Help

```sh
fssh

.> help
Usage:
  help ([command])
Commands:
  !		    shell escape
  cat		concatenate and print files
  cd		change directory
  cp		copy files
  env		prints or sets environment
  exit		exit fssh
  ls		list directory contents
  pwd		print working directory name
  rm		remove files
```

### Connect s3 and copy to gcs

```sh
fssh s3://[S3-Bucket]/

s3://[S3-Bucket]> ls
dir1/
dir2/
file1.txt
file2.txt

s3://[S3-Bucket]> cp -r dir1 gs://[GCS-Bucket]/
```

## Credentianls

### AWS

fssh tries to use the AWS default authentication for accessing S3 buckets.

See https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html

If you are using non-default AWS authentication, here is an example.

```sh
fssh
./> env AWS_ACCESS_KEY_ID=[your access key]
./> env AWS_SECRET_ACCESS_KEY=[your access secret key]
./> env AWS_REGION=[your region]
./> cd s3://[S3-Bucket]/
s3://[S3-Bucket]>
```

```sh
AWS_PROFILE=custom fssh s3://[S3-Bucket]/
s3://[S3-Bucket]>
```

### Google Cloud

fssh tries to use the Google Cloud default credentials for accessing GCS buckets.

See https://cloud.google.com/docs/authentication/application-default-credentials

If you are using non-default Google credential, here is an example.

```sh
fssh
./> env GOOGLE_APPLICATION_CREDENTIALS=path-to-credential.json
./> cd gs://[GCS-Bucket]
gs://[GCS-Bucket]>
```

```sh
GOOGLE_APPLICATION_CREDENTIALS=path-to-credential.json fssh gs://[GCS-Bucket]
gs://[GCS-Bucket]>
```