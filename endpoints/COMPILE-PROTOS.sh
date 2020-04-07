#!/bin/bash

export ANNOTATIONS="/tmp"

echo "Generating descriptor set for Endpoints."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/flame_models.proto \
	proto/flame_service.proto \
	--include_imports \
    --include_source_info \
    --descriptor_set_out=proto.pb