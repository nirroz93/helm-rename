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
	"errors"
	"io"
	"log"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
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

	flags.BoolVar(&rename.MigrateSecrets, "migrate-secrets", true, "")
	flags.BoolVar(&rename.MigrateResources, "migrate-resouces", true, "")
	flags.BoolVar(&rename.CheckNameOverride, "check-nameoverride-value", false, "")
	flags.BoolVar(&rename.YesToAll, "yes", true, "")
	flags.BoolVar(&rename.DryRun, "dry-run", true, "")

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
	_, err := getRelease.Run(renameOptions.OldReleaseName)
	if err != nil {
		log.Printf("Error: Release \"%s\" doesn't exist", renameOptions.OldReleaseName)
		return err
	}

	_, err = getRelease.Run(renameOptions.NewReleaseName)
	if err == nil { // Refactor: Prompt verification
		log.Printf("Error: Release with name \"%s\" already exist", renameOptions.NewReleaseName)
		return err
	}

	if renameOptions.MigrateResources {
		log.Printf("Annotating release resources of \"%s\" with annotation \"meta.helm.sh/release-name: %s\".\n", renameOptions.OldReleaseName, renameOptions.NewReleaseName)
	} else {
		if !renameOptions.DryRun {

		}
	}
	if renameOptions.MigrateSecrets {
		log.Printf("Renaming secrets ##########TODO.\n")
		MigrateReleases(*renameOptions)
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
		CreateRelease(renameOptions, releaseObject)
		DeleteRelease(renameOptions, releaseObject)
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
		return err
	}
	return nil
}
