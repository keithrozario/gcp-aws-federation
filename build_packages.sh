#!/bin/bash
set -e

APP_NAME="get-gcp-token"
VERSION="1.0.0"
MAINTAINER="Keith Krozario <keith.krozario@altostrat.com>"
DESCRIPTION="GCP-AWS Federation Token Fetcher"
HOMEPAGE="https://github.com/altostrat/gcp-aws-federation"

# Ensure clean build
rm -rf build dist
mkdir -p build dist

# Architectures to build
ARCHS=("amd64" "arm64")

for ARCH in "${ARCHS[@]}"; do
    echo "Building for linux/${ARCH}..."
    
    # 1. Compile Go Binary
    # CGO_ENABLED=0 creates a static binary (no external dependencies)
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=${ARCH}
    
    BINARY_NAME="get_gcp_token"
    OUTPUT_DIR="build/${APP_NAME}_${VERSION}_${ARCH}"
    
    mkdir -p "${OUTPUT_DIR}/usr/local/bin"
    mkdir -p "${OUTPUT_DIR}/usr/share/man/man1"
    mkdir -p "${OUTPUT_DIR}/DEBIAN"
    
    go build -o "${OUTPUT_DIR}/usr/local/bin/${BINARY_NAME}" src/get_gcp_token.go
    
    # 2. Copy Man Page
    cp get_gcp_token.1 "${OUTPUT_DIR}/usr/share/man/man1/"
    gzip "${OUTPUT_DIR}/usr/share/man/man1/get_gcp_token.1"
    
    # 3. Create Control File
    # Installed-Size is in kilobytes
    INSTALLED_SIZE=$(du -s "${OUTPUT_DIR}/usr" | cut -f1)
    
    cat <<EOF > "${OUTPUT_DIR}/DEBIAN/control"
Package: ${APP_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Depends: libc6 (>= 2.17)
Maintainer: ${MAINTAINER}
Description: ${DESCRIPTION}
 Fetches OIDC ID tokens from GCP Metadata Server for AWS Federation.
Homepage: ${HOMEPAGE}
Installed-Size: ${INSTALLED_SIZE}
EOF

    # 4. Build .deb Package
    dpkg-deb --build "${OUTPUT_DIR}" "dist/${APP_NAME}_${VERSION}_${ARCH}.deb"
    
    echo "Created dist/${APP_NAME}_${VERSION}_${ARCH}.deb"
done

echo "Build complete. Packages are in dist/"
