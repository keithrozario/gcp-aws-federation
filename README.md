# GCP-AWS Federation

This repository provides a robust, passwordless way to federate from Google Cloud Platform (GCP) to Amazon Web Services (AWS) using OpenID Connect (OIDC). It leverages GCP's Metadata Server to fetch OIDC ID tokens and uses AWS IAM Web Identity Federation to assume roles without needing long-lived AWS Access Keys.

## Overview

The solution consists of a Go-based binary that runs as a background task (e.g. via Cron) to keep a fresh GCP OIDC token available on disk. AWS SDKs and the CLI are then configured to use this token automatically to acquire temporary AWS credentials.

## Components

- `get_gcp_token.go`: A Go source file that compiles into a standalone binary. It fetches an OIDC ID token from the GCP Metadata Server and writes it to a specified file.
- `get_gcp_token`: The compiled binary.
- `AWS IAM Role`: An AWS IAM Role with a trust policy that allows this GCP Principal to Assume it
- `AWS IAM OIDC Provider`: Configure AWS IAM in the account to trust the GCP OIDC Provider
- `AWS SDK`: Any Official AWS SDK (or CLI) 

## Setup (on GCP machine)

### 1. Install via APT

The easiest way to install is via our APT repository (hosted on GitHub Pages).

```bash
# 1. Trust the GPG Key
curl -fsSL https://keithrozario.github.io/gcp-aws-federation/public.key | sudo gpg --dearmor -o /usr/share/keyrings/gcp-aws-federation.gpg

# 2. Add the Repository
echo "deb [signed-by=/usr/share/keyrings/gcp-aws-federation.gpg] https://keithrozario.github.io/gcp-aws-federation stable main" | sudo tee /etc/apt/sources.list.d/gcp-aws-federation.list

# 3. Install
sudo apt update
sudo apt install get-gcp-token
```

### 2. Configure Cron Job

Set up a cron job to renew the token every 12 minutes. This ensures a valid token is always available for the AWS SDKs.


```bash
# Replace 1000 with your actual UID (find it using 'id -u')
# The tool defaults to writing to /run/user/<UID>/aws_gcp_token
*/12 * * * * /usr/local/bin/get_gcp_token
```

Note: **Enable Lingering**

To allow the cron job to run when you are not logged in (critical for background processes), enable lingering:

```bash
loginctl enable-linger $USER
```

Run `crontab -e` and add:

### 3. Environment Variables

Add the following to your shell configuration (e.g., `~/.zshrc` or `~/.bashrc`) so AWS SDKs can find the token:

```bash
# Point to the location where the cron job writes the token
export AWS_WEB_IDENTITY_TOKEN_FILE=/run/user/$(id -u)/aws_gcp_token
export AWS_ROLE_ARN=<AWS_IAM_ROLE_ARN>
```

### 4. Obtain oAuth Client ID
```bash
gcloud iam service-accounts describe <SA_EMAIL> --format="value(uniqueId)"
```

This retrieves the service account ID that will be important later.

## Setup (on AWS)

### 1. OIDC Provider

Create an OIDC Provider with the following:

* URL: accounts.google.com
* ClientIDs: [sts.amazonaws.com, <SA_NUMERIC_ID>]

###  IAM Trust Policy

The AWS IAM Role must be configured with a trust policy that allows the Google OIDC provider to assume the role.

### `aud` vs `oaud`

In the trust policy conditions, we use 3 specific fields to ensure secure and correct federation:

1.  **`accounts.google.com:aud`**: This field is mapped to the **Client ID** (or unique ID) of the Google Service Account (e.g., `10780644A4890271851399`). By checking this, we ensure that only tokens issued to *your* specific service account can assume this role. This field corresponds to the `azp` field of the Google JWT.
2.  **`accounts.google.com:oaud` (Original Audience)**: When requesting a token from GCP, we specify `sts.amazonaws.com` as the audience. Google populates this requested audience in the `oaud` claim of the token. Matching this ensures the token was specifically intended for AWS STS.
3. `**accounts.google.com:sub**` (Subject): The Subject of the token from GCP. This further limits us to the Service account in question. 

Example Trust Policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "<OIDC_PROVIDER_ARN>"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "accounts.google.com:aud": "107806444890271851399",
                    "accounts.google.com:oaud": "sts.amazonaws.com",
                    "accounts.google.com:sub": "107806444890271851399"
                }
            }
        }
    ]
}
```

## Usage

To Use the federation, you ensure the Environment variables are set correctly. The SDKs automatically redeem the AWS credentials from the Google JWTs for you, this solution requires less code on your behalf. You only need to populate the Google JWT to a place that the SDKs (or CLI) can access.

## Notes

It's good practice to write out the Google JWT to `/run/user/{id}` this ensures the file is written only to tempfs and not the disk.
The SDKs will take care of caching and renewal of the AWS Credentials, less things for you to worry about.
We set the Cronjob to once every 12 minutes, which seems like a good compromise as our JWT tokens usually last 1 hour, you can increase/decrease the frequency as you wish.

