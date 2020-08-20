#!/usr/bin/env bash
# https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

set -o errexit \
    -o pipefail \
    -o nounset

package="${1:-lagrande}"
if [[ -z "${package}" ]]; then
    echo "usage: ${0} <package-name>" "<version-string>"
    exit 1
fi
IFS='/' read -r -a package_split <<< "${package}"

package_name=${package_split[${#package_split[@]}-1]}

if [[ "${package}" = 'lagrande' ]]; then
    package_location='.'
else
    package_location="${package}"
fi

version="${2:-$(cat version)}"
build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

if [[ -z "${version}" ]]; then
    echo "usage: ${0} <package-name>" "<version-string>"
    exit 1
fi

platforms=("darwin/amd64" "linux/amd64" "linux/386")

for platform in "${platforms[@]}"
do
    IFS='/' read -r -a platform_split <<< "${platform}"

    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"

    output_name="${package_name}-${GOOS}-${GOARCH}"

    if [ "${GOOS}" = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS="${GOOS}" GOARCH="${GOARCH}" \
        go build \
        -ldflags "-X main.version=${version} -X main.buildDate=${build_date}" \
        -o "${output_name}" \
        "${package_location}"

    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
