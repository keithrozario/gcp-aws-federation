# Plan: Automate APT Repository Hosting on GitHub Pages

This document outlines the strategy to automate the build, signing, and publishing of the `get-gcp-token` Debian package using GitHub Actions and GitHub Pages.

## Overview

The goal is to have a fully automated pipeline where pushing a new tag (e.g., `v1.0.1`) to the repository triggers:
1.  **Build**: Compilation of the Go binary for `amd64` and `arm64`.
2.  **Package**: Creation of `.deb` files.
3.  **Repo Generation**: Updates the APT repository structure (index files).
4.  **Sign**: GPG signing of the Release file using a secret key.
5.  **Publish**: Deployment of the repository to the `gh-pages` branch.

## Prerequisites

1.  **GPG Key Pair**:
    *   Generate a dedicated GPG key pair for signing packages.
    *   Export the private key (ASCII armored) to a GitHub Secret named `GPG_PRIVATE_KEY`.
    *   Commit the public key (`public.key`) to the repository root.

2.  **GitHub Repository Settings**:
    *   Enable **GitHub Pages** to serve from the `gh-pages` branch.
    *   Configure `GPG_PRIVATE_KEY` in **Settings > Secrets and variables > Actions**.

## Task List

### 1. Repository Preparation
- [ ] **Generate GPG Key**: Create a new key pair specifically for this project (do not use your personal key).
- [ ] **Export Keys**:
    -   `gpg --armor --export-secret-keys <KEY_ID> > private.key` (for GitHub Secret).
    -   `gpg --armor --export <KEY_ID> > public.key` (commit to repo).
- [ ] **Configure GitHub Secrets**: Add `GPG_PRIVATE_KEY` and optionally `GPG_PASSPHRASE` (if the key is protected).

### 2. Workflow Automation (.github/workflows/release.yml)
Create a workflow file that runs on `push: tags: 'v*'`.

**Job Steps:**
1.  **Checkout Code**: Fetch the repository.
2.  **Setup Go**: Install the correct Go version.
3.  **Import GPG Key**:
    ```yaml
    - name: Import GPG Key
      uses: crazy-max/ghaction-import-gpg@v6
      with:
        gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
        passphrase: ${{ secrets.GPG_PASSPHRASE }}
    ```
4.  **Build Packages (Matrix Strategy)**:
    *   Use `matrix: { goos: [linux], goarch: [amd64, arm64] }` to build binaries in parallel.
    *   This replaces the loop in `build_packages.sh` for the CI environment, providing better concurrency and error reporting.
    *   `build_packages.sh` is kept for local testing.
5.  **Generate Repository**:
    *   Modify `generate_repo.sh` to support incremental updates (optional) or full rebuilds.
    *   Ensure the script uses the imported GPG key for signing.
6.  **Deploy to GitHub Pages**:
    *   Use `peaceiris/actions-gh-pages` to push the `gcs-repo` (or `public`) folder to the `gh-pages` branch.

### 3. Documentation Update
- [ ] Update `README.md` or `instructions.md` to reflect the new GitHub Pages URL.
- [ ] Provide the new `curl` command for the public key (e.g., `https://<user>.github.io/<repo>/public.key`).
- [ ] Provide the new `sources.list` entry.

## Execution

Once the workflow is in place:
1.  Tag a release: `git tag v1.0.0`.
2.  Push the tag: `git push origin v1.0.0`.
3.  Watch the Action build and deploy the site.
