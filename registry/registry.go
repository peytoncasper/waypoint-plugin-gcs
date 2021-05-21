package registry

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"google.golang.org/api/option"
	"io/ioutil"
	"os"
)

type RegistryConfig struct {
	Bucket string `hcl:"bucket"`
	ObjectName string `hcl:"object_name"`
	ObjectPath string `hcl:"object_path"`
	FilePath string `hcl:"file_path"`

	CredentialPath string `hcl:"credential_path,optional"`
}

type Registry struct {
	config RegistryConfig
}

// Implement Configurable
func (r *Registry) Config() (interface{}, error) {
	return &r.config, nil
}

// Implement ConfigurableNotify
func (r *Registry) ConfigSet(config interface{}) error {
	c, ok := config.(*RegistryConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("Expected *RegisterConfig as parameter")
	}

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && c.CredentialPath == "" {
		return fmt.Errorf("GCS credentials are not set")
	}

	return nil
}

// Implement Registry
func (r *Registry) PushFunc() interface{} {
	// return a function which will be called by Waypoint
	return r.push
}

// A PushFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - *datadir.Project
// - *datadir.App
// - *datadir.Component
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// In addition to default input parameters the builder.Binary from the Build step
// can also be injected.
//
// The output parameters for PushFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
func (r *Registry) push(ctx context.Context, ui terminal.UI) (*Artifact, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Pushing object to GCS")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(r.config.CredentialPath))
	if err != nil {
		return nil, err
	}

	fiile, err := ioutil.ReadFile(r.config.FilePath)
	objectPath := r.config.ObjectPath + r.config.ObjectName

	obj := client.Bucket(r.config.Bucket).Object(objectPath)

	w := obj.NewWriter(ctx)
	_, err = w.Write(fiile)

	if err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return &Artifact{Bucket: r.config.Bucket, ObjectPath: objectPath}, nil
}

// Implement Authenticator
