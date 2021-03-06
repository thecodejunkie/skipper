// Copyright 2015 Zalando SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/zalando/skipper/eskip"
)

type (
	routeMap map[string]*eskip.Route
)

func any(_ *eskip.Route) bool { return true }

func routesDiffer(left, right *eskip.Route) bool {
	return left.String() != right.String()
}

func mapRoutes(routes []*eskip.Route) routeMap {
	m := make(routeMap)
	for _, r := range routes {
		m[r.Id] = r
	}

	return m
}

// take items from 'routes' that don't exist in 'ref' or are different.
func takeDiff(ref []*eskip.Route, routes []*eskip.Route) []*eskip.Route {
	mref := mapRoutes(ref)
	var diff []*eskip.Route
	for _, r := range routes {
		if rr, exists := mref[r.Id]; !exists || routesDiffer(rr, r) {
			diff = append(diff, r)
		}
	}

	return diff
}

// insert/update routes from 'update' that don't exist in 'existing' or
// are different from the one with the same id in 'existing'.
func upsertDifferent(existing []*eskip.Route, update []*eskip.Route, writeClient writeClient) error {
	diff := takeDiff(existing, update)
	return writeClient.UpsertAll(diff)
}

// command executed for upsert.
func upsertCmd(a cmdArgs) error {
	// take input routes:
	routes, err := loadRoutesChecked(a.in)
	if err != nil {
		return err
	}

	wc, err := createWriteClient(a.out)
	if err != nil {
		return err
	}

	return wc.UpsertAll(routes)
}

// command executed for reset.
func resetCmd(a cmdArgs) error {
	// take input routes:
	routes, err := loadRoutesChecked(a.in)
	if err != nil {
		return err
	}

	// take existing routes from output:
	existing := loadRoutesUnchecked(a.out)

	// upsert routes that don't exist or are different:
	wc, err := createWriteClient(a.out)
	if err != nil {
		return err
	}
	err = upsertDifferent(existing, routes, wc)
	if err != nil {
		return err
	}

	// delete routes from existing that were not upserted:
	rm := mapRoutes(routes)
	notSet := func(r *eskip.Route) bool {
		_, set := rm[r.Id]
		return !set
	}

	return wc.DeleteAllIf(existing, notSet)
}

// command executed for delete.
func deleteCmd(a cmdArgs) error {
	// take input routes:
	routes, err := loadRoutesChecked(a.in)
	if err != nil {
		return err
	}

	// delete them:
	wc, err := createWriteClient(a.out)
	if err != nil {
		return err
	}
	return wc.DeleteAllIf(routes, any)
}
