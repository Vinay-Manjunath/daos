//
// (C) Copyright 2020 Intel Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// GOVERNMENT LICENSE RIGHTS-OPEN SOURCE SOFTWARE
// The Government's rights to use, modify, reproduce, release, perform, display,
// or disclose this software are subject to the terms of the Apache License as
// provided in Contract No. 8F-30005.
// Any reproduction of computer software, computer software documentation, or
// portions thereof marked with this legend must also reproduce the markings.
//

package control

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/daos-stack/daos/src/control/common"
	"github.com/daos-stack/daos/src/control/lib/hostlist"
)

func TestControl_HostErrorsMap(t *testing.T) {
	makeHosts := func(hosts ...string) []string {
		return hosts
	}
	makeErrors := func(errStrings ...string) (errs []error) {
		for _, errStr := range errStrings {
			errs = append(errs, errors.New(errStr))
		}
		return
	}

	for name, tc := range map[string]struct {
		hosts     []string
		errors    []error
		expErrMap HostErrorsMap
		expErr    error
	}{
		"nil host error": {
			hosts:     makeHosts("host1"),
			errors:    []error{nil},
			expErrMap: HostErrorsMap{},
		},
		"one host one error": {
			hosts:  makeHosts("host1"),
			errors: makeErrors("whoops"),
			expErrMap: HostErrorsMap{
				"whoops": mockHostSet(t, "host1"),
			},
		},
		"two hosts one error": {
			hosts:  makeHosts("host1", "host2"),
			errors: makeErrors("whoops", "whoops"),
			expErrMap: HostErrorsMap{
				"whoops": mockHostSet(t, "host1,host2"),
			},
		},
		"two hosts two errors": {
			hosts:  makeHosts("host1", "host2"),
			errors: makeErrors("whoops", "oops"),
			expErrMap: HostErrorsMap{
				"whoops": mockHostSet(t, "host1"),
				"oops":   mockHostSet(t, "host2"),
			},
		},
		"two hosts same port one error": {
			hosts:  makeHosts("host1:1", "host2:1"),
			errors: makeErrors("whoops", "whoops"),
			expErrMap: HostErrorsMap{
				"whoops": mockHostSet(t, "host[1-2]:1"),
			},
		},
		"two hosts different port one error": {
			hosts:  makeHosts("host1:1", "host2:2"),
			errors: makeErrors("whoops", "whoops"),
			expErrMap: HostErrorsMap{
				"whoops": mockHostSet(t, "host1:1,host2:2"),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			hem := make(HostErrorsMap)
			for i, host := range tc.hosts {
				gotErr := hem.Add(host, tc.errors[i])
				common.CmpErr(t, tc.expErr, gotErr)
				if tc.expErr != nil {
					return
				}
			}

			cmpOpts := []cmp.Option{
				cmp.Comparer(func(x, y *hostlist.HostSet) bool {
					return x.RangedString() == y.RangedString()
				}),
			}

			if diff := cmp.Diff(tc.expErrMap, hem, cmpOpts...); diff != "" {
				t.Fatalf("unexpected map (-want, +got):\n%s\n", diff)
			}
		})
	}
}
