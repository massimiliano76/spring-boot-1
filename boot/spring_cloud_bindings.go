/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package boot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"

	_ "github.com/paketo-buildpacks/spring-boot/boot/statik"
)

type SpringCloudBindings struct {
	Dependency       libpak.BuildpackDependency
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
	SpringBootLib    string
}

func NewSpringCloudBindings(springBootLib string, dependency libpak.BuildpackDependency, cache libpak.DependencyCache,
	plan *libcnb.BuildpackPlan) SpringCloudBindings {

	return SpringCloudBindings{
		Dependency:       dependency,
		LayerContributor: libpak.NewDependencyLayerContributor(dependency, cache, plan),
		SpringBootLib:    springBootLib,
	}
}

//go:generate statik -src . -include *.sh

func (s SpringCloudBindings) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	s.LayerContributor.Logger = s.Logger

	file := filepath.Join(layer.Path, filepath.Base(s.Dependency.URI))

	layer, err := s.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		s.Logger.Bodyf("Copying to %s", layer.Path)

		if err := sherpa.CopyFile(artifact, file); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy artifact to %s\n%w", file, err)
		}

		s, err := sherpa.StaticFile("/spring-cloud-bindings.sh")
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load spring-cloud-bindings.sh\n%w", err)
		}
		layer.Profile.Add("spring-cloud-bindings.sh", s)

		layer.Launch = true
		return layer, nil
	})
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to contribute spring-cloud-bindings layer\n%w", err)
	}

	if err := os.MkdirAll(s.SpringBootLib, 0755); err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to create directory %s\n%w", s.SpringBootLib, err)
	}

	target := filepath.Join(s.SpringBootLib, filepath.Base(file))
	if err := os.Symlink(file, target); err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to link %s to %s\n%w", file, target, err)
	}

	return layer, nil
}

func (s SpringCloudBindings) Name() string {
	return s.LayerContributor.LayerName()
}
