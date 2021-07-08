/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
)

var (
	renameHelp = `
		Help!
	`
)

type RenameOptions struct {
	MigrateSecrets    bool
	MigrateResources  bool
	CheckNameOverride bool
	YesToAll          bool // Refactor: Rename
	DryRun            bool
	OldReleaseName    string
	NewReleaseName    string
	cfg               *action.Configuration
}

func newRenameCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	rename := &RenameOptions{}
	cmd := &cobra.Command{
		Use:   "rename [flags] RELEASE",
		Short: "Renames a release",
		Long:  renameHelp,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("This command needs 2 arguments: old release name and new release name")
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			rename.cfg = cfg
			rename.OldReleaseName = args[0]
			rename.NewReleaseName = args[1]
			return rename.Rename()
		},
	}

	flags := cmd.Flags()

	flags.BoolVar(&rename.MigrateSecrets, "migrate-secrets", true, "Migrate helm releases to new names (otherwise migrate only resources)")
	flags.BoolVar(&rename.MigrateResources, "migrate-resources", true, "Annotate existis objects created by the release with owenership annotation of the new release name")
	// flags.BoolVar(&rename.CheckNameOverride, "check-nameoverride-value", false, "") // TODO
	flags.BoolVar(&rename.YesToAll, "yes", false, "Quiet mode, answer yes to any verification (existing release for newName for now")
	flags.BoolVar(&rename.DryRun, "dry-run", true, "Dry run, only print actions that will be taken, don't actually do them")

	return cmd

}

// TODO comment
func (renameOptions *RenameOptions) Rename() error {
	if renameOptions.DryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
	}

	log.Printf("Release \"%s\" will be renamed to \"%s\"\n", renameOptions.OldReleaseName, renameOptions.NewReleaseName)

	getRelease := action.NewGet(renameOptions.cfg)
	// getRelease.Version = 0
	oldReleaseObject, err := getRelease.Run(renameOptions.OldReleaseName)
	if err != nil {
		return errors.Wrapf(err, "Error: Release \"%s\" doesn't exist", renameOptions.OldReleaseName)
	}

	_, err = getRelease.Run(renameOptions.NewReleaseName)
	if err == nil {
		log.Printf("Release with name \"%s\" already exist. Do you want to continue with the operation? Note: THIS IS UNTESTED YET, make sure the revision of old release don't exist for the new release [y/N]", renameOptions.NewReleaseName)
		if !renameOptions.YesToAll && !renameOptions.DryRun {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return errors.Wrap(err, "Couldn't read from standard input")
			}
			answer := scanner.Text()
			if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" {
				return errors.Wrapf(err, "Release with name \"%s\" already exist", renameOptions.NewReleaseName)
			}
		}
		// TODO: Set start version if existing release version>current version
	}

	//TODO: add check for fullNameOverride/Usage of Release.Name in the chart.

	if renameOptions.MigrateResources {
		log.Printf("Annotating release resources of \"%s\" with annotation \"meta.helm.sh/release-name: %s\".\n", renameOptions.OldReleaseName, renameOptions.NewReleaseName)
		if !renameOptions.DryRun {
			target, err := renameOptions.cfg.KubeClient.Build(bytes.NewBufferString(oldReleaseObject.Manifest), false)
			if err != nil {
				return err
			}
			err = target.Visit(setOwnerAnnotationVisitor(*renameOptions))
			if err != nil {
				log.Printf("Annotating object failed")
				return err
			}

		}
	}
	if renameOptions.MigrateSecrets {
		log.Printf("Renaming releases (release Secret objects).\n")
		err := MigrateReleases(*renameOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func MigrateReleases(renameOptions RenameOptions) error {
	getReleaseHistory := action.NewHistory(renameOptions.cfg)
	allReleases, err := getReleaseHistory.Run(renameOptions.OldReleaseName)
	if err != nil {
		return err
	}
	for _, releaseObject := range allReleases {
		err = CreateRelease(renameOptions, releaseObject)
		if err != nil {
			return err
		}
		err = DeleteRelease(renameOptions, releaseObject)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateRelease(renameOptions RenameOptions, releaseObject *release.Release) error {
	// release_regex := regexp.MustCompile(fmt.Sprintf("(sh\\.helm\\.release.v1\\.)%s(\\.v2)", renameOptions.OldReleaseName))
	// newName := release_regex.ReplaceAllString(releaseObject.ObjectMeta.Name, fmt.Sprintf("$1%s$2", renameOptions.NewReleaseName))
	log.Printf("Migrating release \"%s\" version %d to new release name \"%s\".\n", renameOptions.OldReleaseName, releaseObject.Version, renameOptions.NewReleaseName)
	newRelease := &release.Release{
		Name:      renameOptions.NewReleaseName,
		Namespace: releaseObject.Namespace,
		Chart:     releaseObject.Chart,
		Config:    releaseObject.Config,
		Info:      releaseObject.Info,
		Manifest:  releaseObject.Manifest,
		Hooks:     releaseObject.Hooks,
		Version:   releaseObject.Version,
	}
	if !renameOptions.DryRun {
		return renameOptions.cfg.Releases.Create(newRelease)
	}
	return nil
}

func DeleteRelease(renameOptions RenameOptions, releaseObject *release.Release) error {
	log.Printf("Deleting release \"%s\" version %d.\n", renameOptions.OldReleaseName, releaseObject.Version)
	if !renameOptions.DryRun {
		_, err := renameOptions.cfg.Releases.Delete(releaseObject.Name, releaseObject.Version)
		if err != nil {
			log.Printf("Deleting release failed")
			return err
		}
	}
	return nil
}

func setOwnerAnnotationVisitor(renameOptions RenameOptions) resource.VisitorFunc {
	releaseName := renameOptions.NewReleaseName
	return func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		patch := fmt.Sprintf("{\"metadata\":{\"annotations\":{\"meta.helm.sh/release-name\":\"%s\"}}}", releaseName)
		log.Printf("Patching \"%s/%s\" with annotation meta.helm.sh/release-name=%s", info.Mapping.GroupVersionKind.Kind, info.Name, releaseName)
		helper := resource.NewHelper(info.Client, info.Mapping)
		obj, err := helper.Patch(info.Namespace, info.Name, types.StrategicMergePatchType, []byte(patch), nil)
		if err != nil {
			return errors.Wrapf(err, "cannot patch %q with kind %s", info.Name, info.Mapping.GroupVersionKind.Kind)
		}
		err = info.Refresh(obj, true)
		if err != nil {
			return err
		}
		return nil
	}
}
