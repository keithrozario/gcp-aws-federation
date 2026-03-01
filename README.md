# GCP-AWS Federation

This repository provides a robust, passwordless way to federate from Google Cloud Platform (GCP) to Amazon Web Services (AWS) using OpenID Connect (OIDC). It leverages GCP's Metadata Server to fetch OIDC ID tokens and uses AWS IAM Web Identity Federation to assume roles without needing long-lived AWS Access Keys.

## Overview

The solution consists of a Go-based binary that runs as a background task (via Cron) to keep a fresh GCP OIDC token available on disk. AWS SDKs and the CLI are then configured to use this token automatically to acquire temporary AWS credentials.

## Components

- `get_gcp_token.go`: A Go source file that compiles into a standalone binary. It fetches an OIDC ID token from the GCP Metadata Server and writes it to a specified file.
- `get_gcp_token`: The compiled binary.
- `iam-role.json`: An example IAM trust policy for the AWS role.
- `aws_federation.py`: (Legacy/Reference) A Python implementation of the federation logic.

## Setup

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

### Testing Beta Releases (Release Candidates)

To test upcoming features, you can install from our beta channel. Note that these releases may be unstable.

```bash
# 1. Trust the GPG Key (Same as stable)
curl -fsSL https://keithrozario.github.io/gcp-aws-federation/public.key | sudo gpg --dearmor -o /usr/share/keyrings/gcp-aws-federation.gpg

# 2. Add the Beta Repository
# Note the '/beta' suffix in the URL
echo "deb [signed-by=/usr/share/keyrings/gcp-aws-federation.gpg] https://keithrozario.github.io/gcp-aws-federation/beta stable main" | sudo tee /etc/apt/sources.list.d/gcp-aws-federation-beta.list

# 3. Install
sudo apt update
sudo apt install get-gcp-token
```

### 2. Configure Cron Job

Set up a cron job to renew the token every 12 minutes. This ensures a valid token is always available for the AWS SDKs.

**Enable Lingering**

To allow the cron job to run when you are not logged in (critical for background processes), enable lingering:

```bash
loginctl enable-linger $USER
```

Run `crontab -e` and add:

```bash
# Replace 1000 with your actual UID (find it using 'id -u')
# The tool defaults to writing to /run/user/<UID>/aws_gcp_token
*/12 * * * * /usr/local/bin/get_gcp_token
```

### 3. Environment Variables

Add the following to your shell configuration (e.g., `~/.zshrc` or `~/.bashrc`) so AWS SDKs can find the token:

```bash
# Point to the location where the cron job writes the token
export AWS_WEB_IDENTITY_TOKEN_FILE=/run/user/$(id -u)/aws_gcp_token
export AWS_ROLE_ARN=arn:aws:iam::941052394956:role/GCPFederation
export AWS_DEFAULT_REGION=ap-southeast-1
```

## IAM Trust Policy

The AWS IAM Role must be configured with a trust policy that allows the Google OIDC provider to assume the role.

### `aud` vs `oaud`

In the trust policy conditions, we use two specific fields to ensure secure and correct federation:

1.  **`accounts.google.com:aud`**: This field is mapped to the **Client ID** (or unique ID) of the Google Service Account (e.g., `107806444890271851399`). By checking this, we ensure that only tokens issued to *your* specific service account can assume this role.
2.  **`accounts.google.com:oaud` (Original Audience)**: When requesting a token from GCP, we specify `sts.amazonaws.com` as the audience. Google populates this requested audience in the `oaud` claim of the token. Matching this ensures the token was specifically intended for AWS STS.

Example Trust Policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::941052394956:oidc-provider/accounts.google.com"
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

## Documentation

The binary comes with a built-in man page for quick reference.

```bash
man get_gcp_token
```
