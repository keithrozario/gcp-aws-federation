# GCP-AWS Federation: Installation Guide

This repository is hosted on GitHub Pages as an APT repository. Follow these steps to install the `get-gcp-token` tool on your Debian/Ubuntu systems.

## Client Installation Instructions

### Step 1: Trust the Repository GPG Key
Download and add the public key to the system's keyring so `apt` can verify the package signatures.
```bash
curl -fsSL https://keithrozario.github.io/gcp-aws-federation/public.key | sudo gpg --dearmor -o /usr/share/keyrings/gcp-aws-federation.gpg
```

### Step 2: Add the Repository to Sources
Add the GitHub Pages repository as a software source.
```bash
echo "deb [signed-by=/usr/share/keyrings/gcp-aws-federation.gpg] https://keithrozario.github.io/gcp-aws-federation stable main" | sudo tee /etc/apt/sources.list.d/gcp-aws-federation.list
```

### Step 3: Install the Package
Update the package index and install the binary.
```bash
sudo apt update
sudo apt install get-gcp-token
```

---

## Post-Installation Verification

Once installed, users can verify the installation using the following commands:

- **Check Binary**:
  ```bash
  which get_gcp_token
  ```
- **Access Documentation**:
  ```bash
  man get_gcp_token
  ```
- **Run Help**:
  ```bash
  get_gcp_token --help
  ```

---

## Maintenance: Releasing a New Version

The release process is fully automated via GitHub Actions.

1.  **Commit Changes**: Ensure your code changes are committed to `main`.
2.  **Tag the Release**: Create a new git tag starting with `v` (e.g., `v1.0.1`).
    ```bash
    git tag v1.0.1
    git push origin v1.0.1
    ```
3.  **Monitor Build**: Check the "Actions" tab in the GitHub repository. The workflow will:
    -   Build binaries for `amd64` and `arm64`.
    -   Generate the APT repository.
    -   Sign the release with the GPG key (stored in Secrets).
    -   Publish the result to the `gh-pages` branch.

