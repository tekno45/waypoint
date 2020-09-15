package runner

import (
	"context"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/waypoint/internal/core"
	pb "github.com/hashicorp/waypoint/internal/server/gen"
	"github.com/hashicorp/waypoint/sdk/component"
)

func (r *Runner) executeDocsOp(
	ctx context.Context,
	log hclog.Logger,
	job *pb.Job,
	project *core.Project,
) (*pb.Job_Result, error) {
	_, ok := job.Operation.(*pb.Job_Docs)
	if !ok {
		// this shouldn't happen since the call to this function is gated
		// on the above type match.
		panic("operation not expected type")
	}

	app, err := project.App(job.Application.Application)
	if err != nil {
		return nil, err
	}

	var results []*pb.Job_DocsResult_Result

	for _, c := range app.Components() {
		info := app.ComponentProto(c)
		if info == nil {
			// Should never happen
			continue
		}

		L := log.With("type", info.Type.String(), "name", info.Name)
		L.Debug("getting docs")

		docs, err := component.Documentation(c)
		if err != nil {
			return nil, err
		}

		if docs == nil {
			L.Debug("no docs for component", "name", info.Name, "type", hclog.Fmt("%T", c))
			continue
		}

		// Start building our result. We append it right away. Since we're
		// appending a pointer we can keep modifying it.
		var result pb.Job_DocsResult_Result
		results = append(results, &result)
		result.Component = info

		var pbdocs pb.Job_DocsResult_Documentation
		dets := docs.Details()
		pbdocs.Description = dets.Description
		pbdocs.Example = dets.Example
		pbdocs.Fields = make(map[string]*pb.Job_DocsResult_FieldDocumentation)

		fields := docs.Fields()

		L.Debug("docs on component", "fields", len(fields))

		for _, f := range docs.Fields() {
			var pbf pb.Job_DocsResult_FieldDocumentation

			pbf.Name = f.Field
			pbf.Type = f.Type
			pbf.Optional = f.Optional
			pbf.Synopsis = f.Synopsis
			pbf.Summary = f.Summary
			pbf.Default = f.Default
			pbf.EnvVar = f.EnvVar

			pbdocs.Fields[f.Field] = &pbf
		}

		result.Docs = &pbdocs
	}

	return &pb.Job_Result{
		Docs: &pb.Job_DocsResult{
			Results: results,
		},
	}, nil
}
