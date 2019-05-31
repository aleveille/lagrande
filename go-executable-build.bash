#!/usr/bin/env bash
# https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

package=$1
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>" "<version-string>"
  exit 1
fi
package_split=(${package//\// })
package_name=${package_split[${#package_split[@]}-1]}

version=$2
build_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)

if [[ -z "$version" ]]; then
    echo "usage: $0 <package-name>" "<version-string>"
    exit 1
fi

platforms=("darwin/amd64" "linux/amd64" "linux/386")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-X main.version=$version -X main.buildDate=$build_date" -o $output_name $package 
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
