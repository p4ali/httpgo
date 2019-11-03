#!/usr/bin/env bash

if [ $# -lt 1 ]; then
  echo
  echo "Usage: $0 tag"
  echo " e.g.: $0 1.1.1"
  echo
  exit 0
fi
ver=$1

create_new_tag(){
  # assume you get the recent tip, and committed everything
  git fetch origin
  git checkout master
  git merge origin/master

  # update httpgo version when build docker
  if [[ "$OSTYPE" == "linux-gnu" ]]; then
    sed -i "s/VERSION=.*/VERSION=${ver}/g" ./Dockerfile
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i "" "s/VERSION=.*/VERSION=${ver}/g" ./Dockerfile
  fi
  git add Dockerfile
  git commit -m "update httpgo to version ${ver}"

  git tag -f $ver HEAD
  git push origin "v$ver"
}

create_bin_for_macosx() {
  make version=${ver} goos=darwin clean releas e
}

echo "TODO: "
echo "  1) update source and create tag v${ver} and v${ver}-nonroot"
echo "  2) publish tag v${ver} and v${ver}-nonroot"
echo "  3) create binary for macosx - httpgo-${ver}-darwin.tar.gz"
echo "  4) draft a new release v${ver} with Github UI (or use auto https://intuit.github.io/auto/pages/generated/release.html to automate), and upload the binary in 3)"
echo "  5) update the homebrew-tools config for the new release - 'brew install p4ali/tools/httpgo' will work (do NOT use force to overwrite)"
echo "  6) go to docker hub and add a new build for the tag of v$ver and trigger a build for both image httpbin:${ver} and httpbin:${ver}-nonroot"
echo "  7) revert the local master branch"

