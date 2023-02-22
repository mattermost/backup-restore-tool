#!/bin/bash

# Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
# See LICENSE.txt for license information.

set -eoux

if [ "$TAG" = "" ]; then
    echo "TAG was not provided"
    exit 1
fi

echo $DOCKERHUB_TOKEN | docker login --username $DOCKERHUB_USERNAME --password-stdin

echo "Tagging images with SHA $TAG"
docker tag mattermost/backup-restore-tool:test mattermost/backup-restore-tool:$TAG
docker push mattermost/backup-restore-tool:$TAG

if [ "$REF_NAME" = "master" ] || [ "$REF_NAME" = "main" ]; then
    echo "Tagging images with 'latest' tag"

    docker tag mattermost/backup-restore-tool:test mattermost/backup-restore-tool:latest
    docker push mattermost/backup-restore-tool:latest
fi
