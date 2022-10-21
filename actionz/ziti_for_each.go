/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package actionz

import (
	"github.com/openziti/runzmd"
	"github.com/pkg/errors"
	"strconv"
)

type ZitiForEach struct{}

func (self *ZitiForEach) Execute(ctx *runzmd.ActionContext) error {
	ctx.Headers["templatize"] = "true"
	apiType, found := ctx.Headers["api"]
	if !found {
		apiType = "edge"
	}
	entityType := ctx.Headers["type"]

	filter, _ := ctx.Headers["filter"]

	minCount := 1
	if count, found := ctx.Headers["minCount"]; found {
		val, err := strconv.Atoi(count)
		if err != nil {
			return errors.Wrapf(err, "couldn't parse minCount, invalid value '%v'", count)
		}
		if val < 0 {
			return errors.Wrapf(err, "invalid minCount, invalid value '%v', must >= 0", count)
		}
		minCount = val
	}

	maxCount := 1
	if count, found := ctx.Headers["maxCount"]; found {
		val, err := strconv.Atoi(count)
		if err != nil {
			return errors.Wrapf(err, "couldn't parse maxCount, invalid value '%v'", count)
		}
		if val < minCount {
			return errors.Wrapf(err, "invalid maxCount '%v', must >= minCount of %v", count, minCount)
		}
		maxCount = val
	}

	entities, err := zitiList(apiType, "list", entityType, "-j", filter)
	if err != nil {
		return err
	}

	if len(entities) < minCount {
		return errors.Errorf("expected at least %v %v, only found %v", minCount, entityType, len(entities))
	}

	if len(entities) > maxCount {
		return errors.Errorf("expected at most %v %v, only found %v", maxCount, entityType, len(entities))
	}

	runner := ZitiRunnerAction{}
	originalBody := ctx.Body
	for _, entity := range entities {
		wrapper := wrapGabs(entity)
		id := wrapper.String("id")
		name := wrapper.String("name")

		ctx.Runner.AddVariable("entityId", id)
		ctx.Runner.AddVariable("entityName", name)
		if err := runner.Execute(ctx); err != nil {
			return err
		}
		ctx.Runner.ClearVariable("entityId")
		ctx.Runner.ClearVariable("entityName")
		ctx.Body = originalBody
	}

	return nil
}
