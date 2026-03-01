#!/bin/bash
set -e

REPO_ROOT="gcs-repo"
DIST_CODENAME="stable"
COMPONENT="main"
ARCHS=("amd64" "arm64")

# clean up previous repo
rm -rf "${REPO_ROOT}"
mkdir -p "${REPO_ROOT}/dists/${DIST_CODENAME}/${COMPONENT}/binary-amd64"
mkdir -p "${REPO_ROOT}/dists/${DIST_CODENAME}/${COMPONENT}/binary-arm64"
mkdir -p "${REPO_ROOT}/pool/${COMPONENT}"

# Copy .deb files to pool
cp dist/*.deb "${REPO_ROOT}/pool/${COMPONENT}/"

# Generate Packages.gz
cd "${REPO_ROOT}"
for ARCH in "${ARCHS[@]}"; do
    echo "Scanning packages for ${ARCH}..."
    dpkg-scanpackages --arch ${ARCH} "pool/${COMPONENT}" /dev/null > "dists/${DIST_CODENAME}/${COMPONENT}/binary-${ARCH}/Packages"
    gzip -k -f "dists/${DIST_CODENAME}/${COMPONENT}/binary-${ARCH}/Packages"
done

# Generate Release file
cd "dists/${DIST_CODENAME}"
cat <<EOF > Release
Origin: GCP AWS Federation
Label: GCP-AWS-Federation
Suite: ${DIST_CODENAME}
Codename: ${DIST_CODENAME}
Version: 1.0
Architectures: ${ARCHS[@]}
Components: ${COMPONENT}
Description: GCP to AWS Federation Tools
Date: $(date -R)
EOF

# Calculate hashes for Release file
do_hash() {
    HASH_NAME=$1
    HASH_CMD=$2
    echo "${HASH_NAME}:" >> Release
    find . -type f \( -name "Packages*" -o -name "Release" \) -printf "%P
" | while read -r file; do
        if [ "$file" != "Release" ]; then
            echo " $(${HASH_CMD} "$file" | cut -d" " -f1) $(wc -c < "$file") $file" >> Release
        fi
    done
}

do_hash "MD5Sum" "md5sum"
do_hash "SHA1" "sha1sum"
do_hash "SHA256" "sha256sum"

# Sign Release file
# We need a GPG key. If one doesn't exist, generate a temporary one for this demo.
if ! gpg --list-keys "GCP-AWS-Federation" > /dev/null 2>&1; then
    echo "Generating temporary GPG key for signing..."
    cat > gpg-batch <<EOF
%echo Generating a basic OpenPGP key
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: GCP AWS Federation
Name-Email: build@example.com
Expire-Date: 0
%no-protection
%commit
%echo done
EOF
    gpg --batch --gen-key gpg-batch
    rm gpg-batch
fi

echo "Signing Release file..."
gpg --default-key "GCP AWS Federation" -abs -o Release.gpg Release
gpg --default-key "GCP AWS Federation" --clearsign -o InRelease Release

# Export public key
gpg --armor --export "GCP AWS Federation" > ../../../public.key

echo "Repository generated in ${REPO_ROOT}"
echo "Public key exported to ${REPO_ROOT}/public.key"
