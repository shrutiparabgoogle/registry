// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"log"
	"time"
)

type ResourceCollection struct {
	maxUpdateTime time.Time
	resourceList  []Resource
}

func ProcessManifest(
	ctx context.Context,
	client connection.Client,
	projectID string,
	manifest *rpc.Manifest) ([]string, error) {

	var actions []string
	for _, resource := range manifest.GeneratedResources {

		newActions, err := processManifestResource(ctx, client, projectID, resource)
		if err != nil {
			log.Printf("Skipping resource: %q\nGot error: %s", resource, err.Error())
		}
		actions = append(actions, newActions...)
	}

	return actions, nil
}

func processManifestResource(
	ctx context.Context,
	client connection.Client,
	projectID string,
	resource *rpc.GeneratedResource) ([]string, error) {
	// Generate dependency map
	resourcePattern := fmt.Sprintf("projects/%s/%s", projectID, resource.Pattern)
	dependencyMaps := make([]map[string]ResourceCollection, 0, len(resource.Dependencies))
	for _, d := range resource.Dependencies {
		dMap, err := generateDependencyMap(ctx, client, resourcePattern, d.Pattern, d.Filter)
		if err != nil {
			log.Printf("Error while generating dependency map.\n Error: %s\n Skipping resource %+v", err.Error(), resource)
			continue
		}
		dependencyMaps = append(dependencyMaps, dMap)
	}

	// Generate resource list
	resourceList, err := ListResources(ctx, client, resourcePattern, resource.Filter)
	if err != nil {
		return nil, err
	}

	// Update target resources
	cmds, err := generateActions(ctx, client, resourceList,
		resource.Dependencies, dependencyMaps, resource.Action)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func generateDependencyMap(
	ctx context.Context,
	client connection.Client,
	resourcePattern,
	dependencyPattern,
	dependencyFilter string) (map[string]ResourceCollection, error) {

	sourceMap := make(map[string]ResourceCollection)

	// Extend the source pattern if it contains $resource.api like pattern
	extDependencyPattern, err := ExtendSourcePattern(resourcePattern, dependencyPattern)
	if err != nil {
		return nil, err
	}

	// Fetch resources using the extDependencyPattern
	sourceList, err := ListResources(ctx, client, extDependencyPattern, dependencyFilter)
	if err != nil {
		return nil, err
	}

	for _, source := range sourceList {
		group, err := ExtractGroup(dependencyPattern, source)
		if err != nil {
			return nil, err
		}

		sourceTime := source.GetUpdateTimestamp()
		collection, exists := sourceMap[group]
		if !exists {
			collection = ResourceCollection{
				maxUpdateTime: sourceTime,
			}
		} else if collection.maxUpdateTime.Before(sourceTime) {
			collection.maxUpdateTime = sourceTime
		}

		collection.resourceList = append(collection.resourceList, source)
		sourceMap[group] = collection
	}

	return sourceMap, nil

}

func generateActions(
	ctx context.Context,
	client connection.Client,
	resourceList []Resource,
	dependencies []*rpc.Dependency,
	dependencyMaps []map[string]ResourceCollection,
	action string) ([]string, error) {

	visited := make(map[string]bool, 0)
	cmds := make([]string, 0)

	for _, resource := range resourceList {
		resourceTime := resource.GetUpdateTimestamp()

		takeAction := false
		var args []Resource

		// Evaluate this resource against each dependency source pattern
		for i, dependency := range dependencies {
			dMap := dependencyMaps[i]
			// Get the group to look for in dependencyMap
			group, err := ExtractGroup(dependency.Pattern, resource)
			if err != nil {
				return nil, fmt.Errorf("Cannot match resource with dependency. Error: %s", err.Error())
			}

			if collection, ok := dMap[group]; ok {
				// Take action if dependency timestamp is later than resource timestamp
				if collection.maxUpdateTime.After(resourceTime) {
					takeAction = true
				}
				visited[group] = true
				// TODO: Evaluate if append only the group or resource name should be enough
				args = append(args, collection.resourceList[0])
			} else {
				// For a given resource, each of it's defined dependency group should be present.
				// If any one of the dependency groups is missing, avoid calculating any action for the resource
				takeAction = false
				break
			}
		}

		if takeAction {
			cmd, err := GenerateCommand(action, args)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		}
	}

	// Iterate over first dependency source and evaluate that against remaining dependencies
	if len(dependencyMaps) > 0 {
		dMap0 := dependencyMaps[0]
		for key := range dMap0 {
			takeAction := true
			var args []Resource
			if _, ok := visited[key]; !ok {
				for _, dMap := range dependencyMaps {
					collection, ok := dMap[key]
					if ok {
						args = append(args, collection.resourceList[0])
					} else {
						takeAction = false
						break
					}
				}
			} else {
				takeAction = false
			}

			if takeAction {
				cmd, err := GenerateCommand(action, args)
				if err != nil {
					return nil, err
				}
				cmds = append(cmds, cmd)
			}

		}
	}

	return cmds, nil
}
