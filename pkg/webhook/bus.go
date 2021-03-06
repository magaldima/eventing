/*
 * Copyright 2018 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhook

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	"github.com/mattbaird/jsonpatch"
	"k8s.io/apimachinery/pkg/util/validation"
)

var (
	errInvalidBusInput = errors.New("failed to convert input into Bus or ClusterBus")
	errInternalNilBus  = errors.New("unexpected internal error: nil Bus or ClusterBus")
)

// ValidateBus is Bus resource specific validation and mutation handler
func ValidateBus(ctx context.Context) ResourceCallback {
	return func(patches *[]jsonpatch.JsonPatchOperation, old GenericCRD, new GenericCRD) error {
		oldBus, newBus, err := unmarshalBuses(ctx, old, new, "ValidateBus")
		if err != nil {
			return err
		}

		return validateBus(oldBus, newBus)
	}
}

func validateBus(old, new v1alpha1.GenericBus) error {
	if new.GetSpec().Parameters != nil {
		if new.GetSpec().Parameters.Channel != nil {
			for _, p := range *new.GetSpec().Parameters.Channel {
				errs := validation.IsConfigMapKey(p.Name)
				if len(errs) > 0 {
					return fmt.Errorf("invalid parameter name Spec.Parameters.Channel.%s: %s", p.Name,
						strings.Join(errs, ", "))
				}
			}
		}
		if new.GetSpec().Parameters.Subscription != nil {
			for _, p := range *new.GetSpec().Parameters.Subscription {
				errs := validation.IsConfigMapKey(p.Name)
				if len(errs) > 0 {
					return fmt.Errorf("invalid parameter name Spec.Parameters.Subscription.%s: %s", p.Name,
						strings.Join(errs, ", "))
				}
			}
		}
	}
	return nil
}

func unmarshalBuses(
	ctx context.Context, old, new GenericCRD, fnName string) (v1alpha1.GenericBus, v1alpha1.GenericBus, error) {
	var oldBus v1alpha1.GenericBus
	if old != nil {
		var ok bool
		oldBus, ok = old.(v1alpha1.GenericBus)
		if !ok {
			return nil, nil, errInvalidBusInput
		}
	}
	glog.Infof("%s: OLD Bus is\n%+v", fnName, oldBus)

	if new == nil {
		return nil, nil, errInternalNilBus
	}
	newBus, ok := new.(v1alpha1.GenericBus)
	if !ok {
		return nil, nil, errInvalidBusInput
	}
	glog.Infof("%s: NEW Bus is\n%+v", fnName, newBus)

	return oldBus, newBus, nil
}
